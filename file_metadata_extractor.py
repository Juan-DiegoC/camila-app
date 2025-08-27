#!/usr/bin/env python3

import os
import sys
import argparse
from datetime import datetime
from pathlib import Path
import re
import logging
import traceback
import mimetypes
import csv
import stat
import tempfile
from typing import List, Tuple, Optional

try:
    import PyPDF2
except ImportError:
    print("PyPDF2 not found. Install with: pip install PyPDF2==3.0.1")
    sys.exit(1)

try:
    import fitz  # PyMuPDF
    PYMUPDF_AVAILABLE = True
except ImportError:
    PYMUPDF_AVAILABLE = False
    print("Warning: PyMuPDF not available. PDF processing may be limited for corrupted files.")

# Setup logging
def setup_logging(debug: bool = False):
    """Setup logging configuration."""
    level = logging.DEBUG if debug else logging.INFO
    format_str = '%(asctime)s - %(levelname)s - %(message)s'
    logging.basicConfig(level=level, format=format_str)
    return logging.getLogger(__name__)

logger = setup_logging()

try:
    from openpyxl import load_workbook, Workbook
    from openpyxl.utils import get_column_letter
    from openpyxl.styles import Font, Alignment, Border, Side
except ImportError:
    print("openpyxl not found. Install with: pip install openpyxl==3.1.2")
    sys.exit(1)


def extract_number_from_filename(filename: str) -> int:
    """Extract the number from filenames like '01FileName', '02OtherFileName', etc."""
    match = re.match(r'^(\d+)', filename)
    return int(match.group(1)) if match else float('inf')


def get_file_mime_type(file_path: str) -> str:
    """Get file MIME type using multiple methods."""
    try:
        # First try mimetypes module
        mime_type, _ = mimetypes.guess_type(file_path)
        if mime_type:
            logger.debug(f"MIME type for {file_path}: {mime_type}")
            return mime_type
        
        # Fallback: read file header
        with open(file_path, 'rb') as file:
            header = file.read(16)
            
        # Common file signatures
        signatures = {
            b'%PDF-': 'application/pdf',
            b'\x89PNG': 'image/png',
            b'\xff\xd8\xff': 'image/jpeg',
            b'GIF8': 'image/gif',
            b'RIFF': 'audio/wav',  # or video
            b'ID3': 'audio/mp3',
            b'\xff\xfb': 'audio/mp3',
            b'\xff\xf3': 'audio/mp3',
            b'\xff\xf2': 'audio/mp3',
            b'ftyp': 'video/mp4',  # offset 4 bytes
            b'\x1a\x45\xdf\xa3': 'video/webm',
        }
        
        for sig, mime in signatures.items():
            if header.startswith(sig):
                logger.debug(f"Detected {mime} for {file_path} by signature")
                return mime
                
        # Check MP4 at offset 4
        if len(header) >= 8 and header[4:8] == b'ftyp':
            logger.debug(f"Detected video/mp4 for {file_path} by offset signature")
            return 'video/mp4'
            
        logger.debug(f"Unknown file type for {file_path}, header: {header[:8]}")
        return 'application/octet-stream'
        
    except Exception as e:
        logger.error(f"Error detecting MIME type for {file_path}: {e}")
        return 'application/octet-stream'

def is_valid_pdf(file_path: str) -> bool:
    """Check if file is a valid PDF by reading first few bytes and MIME type."""
    try:
        mime_type = get_file_mime_type(file_path)
        is_pdf_mime = mime_type == 'application/pdf'
        
        with open(file_path, 'rb') as file:
            header = file.read(5)
            is_pdf_header = header == b'%PDF-'
            
        result = is_pdf_mime and is_pdf_header
        logger.debug(f"PDF validation for {file_path}: MIME={is_pdf_mime}, Header={is_pdf_header}, Result={result}")
        return result
        
    except Exception as e:
        logger.error(f"Error validating PDF {file_path}: {e}")
        return False

def get_pdf_pages_pypdf2(file_path: str) -> Optional[int]:
    """Get PDF pages using PyPDF2."""
    try:
        logger.debug(f"Trying PyPDF2 for {file_path}")
        with open(file_path, 'rb') as file:
            pdf_reader = PyPDF2.PdfReader(file)
            page_count = len(pdf_reader.pages)
            logger.debug(f"PyPDF2 success: {page_count} pages for {file_path}")
            return page_count
    except Exception as e:
        logger.error(f"PyPDF2 failed for {file_path}: {type(e).__name__}: {str(e)}")
        logger.debug(f"PyPDF2 full traceback for {file_path}:\n{traceback.format_exc()}")
        return None

def get_pdf_pages_pymupdf(file_path: str) -> Optional[int]:
    """Get PDF pages using PyMuPDF (fallback)."""
    if not PYMUPDF_AVAILABLE:
        logger.debug("PyMuPDF not available")
        return None
    try:
        logger.debug(f"Trying PyMuPDF for {file_path}")
        doc = fitz.open(file_path)
        page_count = len(doc)
        doc.close()
        logger.debug(f"PyMuPDF success: {page_count} pages for {file_path}")
        return page_count
    except Exception as e:
        logger.error(f"PyMuPDF failed for {file_path}: {type(e).__name__}: {str(e)}")
        logger.debug(f"PyMuPDF full traceback for {file_path}:\n{traceback.format_exc()}")
        return None

def get_pdf_pages_estimate(file_path: str) -> int:
    """Estimate PDF pages by searching for page objects in raw content."""
    try:
        logger.debug(f"Trying estimation method for {file_path}")
        with open(file_path, 'rb') as file:
            content = file.read(min(1024*1024, os.path.getsize(file_path)))  # Read max 1MB
            
        # Count occurrences of page object patterns
        page_patterns = [b'/Type /Page', b'/Type/Page', b'endobj']
        pattern_counts = {}
        
        for pattern in page_patterns:
            count = content.count(pattern)
            if pattern == b'endobj':
                count = max(1, count // 10)  # Rough estimate
            pattern_counts[pattern] = count
            
        max_count = max(pattern_counts.values()) if pattern_counts else 1
        estimated_pages = max(1, min(max_count, 1000))  # Cap at reasonable number
        
        logger.debug(f"Pattern counts for {file_path}: {pattern_counts}")
        logger.debug(f"Estimated {estimated_pages} pages for {file_path}")
        return estimated_pages
        
    except Exception as e:
        logger.error(f"Estimation failed for {file_path}: {type(e).__name__}: {str(e)}")
        return 1

def should_process_as_pdf(file_path: str) -> bool:
    """Determine if file should be processed as PDF based on extension AND content."""
    # Only check files with .pdf extension
    if not file_path.lower().endswith('.pdf'):
        return False
        
    # Check if it's actually a PDF file
    mime_type = get_file_mime_type(file_path)
    is_actually_pdf = is_valid_pdf(file_path)
    
    logger.debug(f"PDF check for {file_path}: extension=.pdf, mime={mime_type}, valid_pdf={is_actually_pdf}")
    
    if not is_actually_pdf:
        logger.warning(f"File {file_path} has .pdf extension but is not a valid PDF (MIME: {mime_type})")
        return False
        
    return True

def get_pdf_pages(file_path: str) -> int:
    """Get number of pages in a PDF file with comprehensive error handling."""
    logger.info(f"Processing PDF: {file_path}")
    
    # First check if we should even try to process as PDF
    if not should_process_as_pdf(file_path):
        logger.warning(f"Skipping PDF processing for {file_path} - not a valid PDF")
        return 1
    
    # Try PyPDF2 first
    try:
        pages = get_pdf_pages_pypdf2(file_path)
        if pages is not None:
            logger.info(f"Successfully extracted {pages} pages from {file_path} using PyPDF2")
            return pages
    except Exception as e:
        logger.error(f"Unexpected error in PyPDF2 processing for {file_path}: {e}")
    
    # Try PyMuPDF as fallback
    try:
        pages = get_pdf_pages_pymupdf(file_path)
        if pages is not None:
            logger.info(f"Successfully extracted {pages} pages from {file_path} using PyMuPDF (fallback)")
            return pages
    except Exception as e:
        logger.error(f"Unexpected error in PyMuPDF processing for {file_path}: {e}")
    
    # Last resort: estimate based on file content
    try:
        pages = get_pdf_pages_estimate(file_path)
        logger.warning(f"Using estimation method for {file_path}: {pages} pages")
        return pages
    except Exception as e:
        logger.error(f"All PDF processing methods failed for {file_path}: {e}")
        logger.error(f"Final fallback traceback:\n{traceback.format_exc()}")
        return 1


def get_file_size(file_path: str) -> str:
    """Get file size formatted as KB or MB."""
    try:
        size_bytes = os.path.getsize(file_path)
        if size_bytes < 1024:
            return f"{size_bytes} bytes"
        elif size_bytes < 1024 * 1024:
            size_kb = size_bytes / 1024
            return f"{size_kb:.1f} KB".replace('.', ',')
        else:
            size_mb = size_bytes / (1024 * 1024)
            return f"{size_mb:.2f} MB".replace('.', ',')
    except Exception:
        return "0 bytes"


def count_files_in_directory(dir_path: str) -> int:
    """Count number of files in a directory."""
    try:
        return len([f for f in os.listdir(dir_path) if os.path.isfile(os.path.join(dir_path, f))])
    except Exception:
        return 0


def get_creation_date(file_path: str) -> str:
    """Get file creation date formatted as Spanish date."""
    try:
        stat = os.stat(file_path)
        # Use birth time if available (macOS), otherwise use modification time
        creation_time = getattr(stat, 'st_birthtime', stat.st_mtime)
        dt = datetime.fromtimestamp(creation_time)
        # Format as Spanish date: d/mm/yyyy h:mm a. m./p. m.
        am_pm = "a. m." if dt.hour < 12 else "p. m."
        hour_12 = dt.hour if dt.hour <= 12 else dt.hour - 12
        if hour_12 == 0:
            hour_12 = 12
        return f"{dt.day}/{dt.month:02d}/{dt.year} {hour_12}:{dt.minute:02d} {am_pm}"
    except Exception:
        dt = datetime.now()
        am_pm = "a. m." if dt.hour < 12 else "p. m."
        hour_12 = dt.hour if dt.hour <= 12 else dt.hour - 12
        if hour_12 == 0:
            hour_12 = 12
        return f"{dt.day}/{dt.month:02d}/{dt.year} {hour_12}:{dt.minute:02d} {am_pm}"


def get_ordered_files(directory: str) -> List[Tuple[str, str]]:
    """Get files and directories ordered by their numeric prefix."""
    items = []
    
    for item in os.listdir(directory):
        item_path = os.path.join(directory, item)
        items.append((item, item_path))
    
    # Sort by the numeric prefix extracted from filename
    items.sort(key=lambda x: extract_number_from_filename(x[0]))
    
    return items


def setup_excel_formatting():
    """Setup Excel formatting styles."""
    # Header font and style
    header_font = Font(name='Calibri', size=11, bold=True)
    header_alignment = Alignment(horizontal='center', vertical='center', wrap_text=True)
    
    # Data font and style
    data_font = Font(name='Calibri', size=11)
    data_alignment = Alignment(vertical='center', wrap_text=True)
    
    # Border style
    thin_border = Border(
        left=Side(style='thin'),
        right=Side(style='thin'),
        top=Side(style='thin'),
        bottom=Side(style='thin')
    )
    
    return header_font, header_alignment, data_font, data_alignment, thin_border

def add_headers_and_formatting(ws, directory: str = "", litigant_name: str = ""):
    """Add headers and formatting to match the template."""
    header_font, header_alignment, data_font, data_alignment, thin_border = setup_excel_formatting()
    
    # Add document header information to match indice de ejemplo format
    # Row 1: Ciudad
    ws['A1'].value = 'Ciudad'
    ws['B1'].value = 'SANTIAGO DE CALI (VALLE)'
    ws['H1'].value = 'EXPEDIENTE FÍSICO'
    
    # Row 2: Despacho Judicial (Judge)
    ws['A2'].value = 'Despacho Judicial'
    ws['B2'].value = 'JUZGADO DECIMO LABORALDEL  CIRCUITO DE CALI'  # Standard judge format
    ws['H2'].value = 'El expediente judicial posee documentos físicos:'
    ws['J2'].value = 'SI     NO X'
    
    # Row 3: Serie
    ws['A3'].value = 'Serie o Subserie Documental'
    ws['B3'].value = 'ORDINARIO LABORAL DE PRIMERA INSTANCIA'
    
    # Row 4: Radicación  
    ws['A4'].value = 'No. Radicación del Proceso'
    ws['B4'].value = '76001310501020230040100'  # Standard format, could be made configurable
    ws['H4'].value = 'No. de carpetas (cuadernos), legajos o tomos:'
    
    # Row 5: Demandado
    ws['A5'].value = 'Partes Procesales (Parte A)\n(demandado, procesado, accionado)'
    ws['B5'].value = 'PORVENIR Y OTROS'  # Standard format
    ws['H5'].value = 'No. de carpetas (cuadernos), legajos o tomos digitalizados:'
    
    # Row 6: Demandante (Litigant)
    ws['A6'].value = 'Partes Procesales (Parte B)\n(demandante, denunciante, accionante)'
    ws['B6'].value = litigant_name.upper() if litigant_name else 'NOMBRE DEL LITIGANTE'
    
    # Row 7: Terceros
    ws['A7'].value = 'Terceros Intervinientes'
    
    # Row 8: Cuaderno
    ws['A8'].value = 'Cuaderno '
    
    # Row 10: Main title
    ws['A10'].value = 'ÍNDICE ELECTRÓNICO DEL EXPEDIENTE JUDICIAL'
    
    # Headers in Spanish (row 11)
    headers = {
        'A11': 'Nombre Documento',
        'B11': 'Fecha Creación Documento', 
        'C11': 'Fecha Incorporación Expediente',
        'D11': 'Orden Documento',
        'E11': 'Número Páginas',
        'F11': 'Página Inicio',
        'G11': 'Página Fin',
        'H11': 'Formato',
        'I11': 'Tamaño',
        'J11': 'Origen',
        'K11': 'Observaciones'
    }
    
    # Add headers with formatting
    for cell_ref, header_text in headers.items():
        cell = ws[cell_ref]
        cell.value = header_text
        cell.font = header_font
        cell.alignment = header_alignment
        cell.border = thin_border

def check_file_permissions(file_path: str) -> bool:
    """Check if we can write to the specified file path."""
    try:
        # Check if file exists and is writable
        if os.path.exists(file_path):
            return os.access(file_path, os.W_OK)
        
        # Check if directory is writable
        directory = os.path.dirname(file_path) or '.'
        return os.access(directory, os.W_OK)
    except Exception as e:
        logger.error(f"Error checking permissions for {file_path}: {e}")
        return False

def create_safe_filename(base_name: str) -> str:
    """Create a safe filename, adding timestamp if needed."""
    if check_file_permissions(base_name):
        return base_name
    
    # Try with timestamp
    name, ext = os.path.splitext(base_name)
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    timestamped_name = f"{name}_{timestamp}{ext}"
    
    if check_file_permissions(timestamped_name):
        logger.info(f"Using timestamped filename: {timestamped_name}")
        return timestamped_name
    
    # Fall back to temp directory
    temp_dir = tempfile.gettempdir()
    temp_name = os.path.join(temp_dir, os.path.basename(timestamped_name))
    logger.warning(f"Using temp directory: {temp_name}")
    return temp_name

def export_to_csv(data: List[dict], csv_file: str) -> bool:
    """Export data to CSV format."""
    try:
        logger.info(f"Exporting to CSV: {csv_file}")
        
        # Create safe filename
        safe_csv_file = create_safe_filename(csv_file)
        
        with open(safe_csv_file, 'w', newline='', encoding='utf-8') as file:
            if not data:
                logger.warning("No data to export to CSV")
                return False
                
            writer = csv.DictWriter(file, fieldnames=data[0].keys())
            writer.writeheader()
            writer.writerows(data)
            
        logger.info(f"CSV export successful: {safe_csv_file}")
        print(f"CSV file created: {safe_csv_file}")
        return True
        
    except Exception as e:
        logger.error(f"CSV export failed: {e}")
        logger.error(f"CSV export traceback:\n{traceback.format_exc()}")
        return False

def process_files_to_excel(directory: str, output_file: str = "metadata_output.xlsx", export_csv: bool = False, litigant_name: str = ""):
    """Process files and write metadata to Excel file matching the template format."""
    
    # Get ordered files
    ordered_items = get_ordered_files(directory)
    
    # Create new workbook (always start fresh to match template)
    wb = Workbook()
    ws = wb.active
    ws.title = "Indice Electrónico"
    
    # Add headers and formatting
    add_headers_and_formatting(ws, directory, litigant_name)
    
    header_font, header_alignment, data_font, data_alignment, thin_border = setup_excel_formatting()
    
    # Starting row for data (A12 as specified)
    start_row = 12
    current_page = 1  # Track page numbering
    
    for idx, (item_name, item_path) in enumerate(ordered_items):
        current_row = start_row + idx
        
        # Column A: Nombre Documento (File name)
        ws[f'A{current_row}'].value = item_name
        
        # Column B: Fecha Creación Documento (Creation date)
        creation_date_str = get_creation_date(item_path)
        ws[f'B{current_row}'].value = creation_date_str
        
        # Column C: Fecha Incorporación Expediente (same as creation for now)
        ws[f'C{current_row}'].value = creation_date_str
        
        # Column D: Orden Documento (sequential order)
        ws[f'D{current_row}'].value = idx + 1
        
        if os.path.isfile(item_path):
            # Column E: Número Páginas (Number of pages)
            try:
                if item_path.lower().endswith('.pdf'):
                    logger.info(f"Processing potential PDF: {item_path}")
                    pages = get_pdf_pages(item_path)
                else:
                    pages = 1
                    logger.debug(f"Non-PDF file: {item_path}, setting pages = 1")
                ws[f'E{current_row}'].value = pages
            except Exception as e:
                logger.error(f"Error processing file {item_path} for page count: {e}")
                logger.error(f"Stack trace:\n{traceback.format_exc()}")
                ws[f'E{current_row}'].value = 1  # Fallback
            
            # Column F: Página Inicio (Starting page)
            ws[f'F{current_row}'].value = current_page
            
            # Column G: Página Fin (Ending page) - Formula
            ws[f'G{current_row}'].value = f"=F{current_row}+E{current_row}-1"
            
            # Column H: Formato (File format)
            file_ext = os.path.splitext(item_path)[1].upper().lstrip('.')
            ws[f'H{current_row}'].value = file_ext if file_ext else "UNKNOWN"
            
            # Column I: Tamaño (File size)
            file_size = get_file_size(item_path)
            ws[f'I{current_row}'].value = file_size
            
            # Column J: Origen (Origin)
            ws[f'J{current_row}'].value = "ELECTRONICO"
            
            # Update page counter for next file
            current_page += pages
            
        elif os.path.isdir(item_path):
            # For directories
            file_count = count_files_in_directory(item_path)
            ws[f'E{current_row}'].value = 1  # Directories count as 1 page
            ws[f'F{current_row}'].value = current_page
            ws[f'G{current_row}'].value = current_page  # Same start and end for directories
            ws[f'H{current_row}'].value = "CARPETA"
            ws[f'I{current_row}'].value = f"{file_count} archivos"
            ws[f'J{current_row}'].value = "ELECTRONICO"
            
            current_page += 1
        
        # Apply formatting to all data cells
        for col in ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K']:
            cell = ws[f'{col}{current_row}']
            cell.font = data_font
            cell.alignment = data_alignment
            cell.border = thin_border
    
    # Adjust column widths
    column_widths = {
        'A': 25,  # Nombre Documento
        'B': 18,  # Fecha Creación
        'C': 18,  # Fecha Incorporación  
        'D': 8,   # Orden
        'E': 10,  # Número Páginas
        'F': 10,  # Página Inicio
        'G': 10,  # Página Fin
        'H': 12,  # Formato
        'I': 15,  # Tamaño
        'J': 12,  # Origen
        'K': 20   # Observaciones
    }
    
    for col, width in column_widths.items():
        ws.column_dimensions[col].width = width
    
    # Prepare data for CSV export if requested
    csv_data = []
    if export_csv:
        for idx, (item_name, item_path) in enumerate(ordered_items):
            current_row = start_row + idx
            
            row_data = {
                'Nombre Documento': ws[f'A{current_row}'].value,
                'Fecha Creación Documento': ws[f'B{current_row}'].value,
                'Fecha Incorporación Expediente': ws[f'C{current_row}'].value,
                'Orden Documento': ws[f'D{current_row}'].value,
                'Número Páginas': ws[f'E{current_row}'].value,
                'Página Inicio': ws[f'F{current_row}'].value,
                'Página Fin': ws[f'G{current_row}'].value,
                'Formato': ws[f'H{current_row}'].value,
                'Tamaño': ws[f'I{current_row}'].value,
                'Origen': ws[f'J{current_row}'].value,
                'Observaciones': ws[f'K{current_row}'].value or ''
            }
            csv_data.append(row_data)
    
    # Save the Excel workbook
    try:
        safe_excel_file = create_safe_filename(output_file)
        logger.info(f"Saving Excel file: {safe_excel_file}")
        wb.save(safe_excel_file)
        print(f"Metadata extracted and saved to {safe_excel_file}")
        print(f"Processed {len(ordered_items)} items from {directory}")
        print(f"Format matches 'indice de ejemplo.xlsm' template")
        
        # Export to CSV if requested
        if export_csv:
            csv_file = safe_excel_file.replace('.xlsx', '.csv').replace('.xlsm', '.csv')
            export_to_csv(csv_data, csv_file)
            
    except PermissionError as e:
        logger.error(f"Permission denied when saving Excel file: {e}")
        print(f"ERROR: Permission denied when saving {output_file}")
        print("Possible causes:")
        print("- File is open in Excel or another program")
        print("- Insufficient write permissions in the directory")
        print("- File is read-only")
        
        # Try to export CSV as fallback
        if export_csv or True:  # Always try CSV as fallback
            print("\nTrying CSV export as fallback...")
            csv_file = output_file.replace('.xlsx', '.csv').replace('.xlsm', '.csv')
            if export_to_csv(csv_data, csv_file):
                print("CSV export successful!")
            else:
                print("CSV export also failed.")
        
        raise  # Re-raise to be caught by main()
        
    except Exception as e:
        logger.error(f"Unexpected error saving Excel file: {e}")
        logger.error(f"Excel save traceback:\n{traceback.format_exc()}")
        print(f"ERROR: Unexpected error saving {output_file}: {e}")
        raise


def main():
    parser = argparse.ArgumentParser(description='Extract file metadata and write to Excel')
    parser.add_argument('--directory', '-d', 
                       default=os.getcwd(),
                       help='Directory to process (default: current directory)')
    parser.add_argument('--output', '-o',
                       default='metadata_output.xlsx',
                       help='Output Excel file (default: metadata_output.xlsx)')
    parser.add_argument('--debug', '-v',
                       action='store_true',
                       help='Enable debug logging')
    parser.add_argument('--log-file',
                       help='Write logs to file instead of console')
    parser.add_argument('--csv', '-c',
                       action='store_true',
                       help='Also export to CSV format')
    parser.add_argument('--csv-only',
                       action='store_true',
                       help='Export only to CSV format (no Excel)')
    parser.add_argument('--litigant', '-l',
                       default='',
                       help='Name of the litigant (demandante) for the header')
    
    args = parser.parse_args()
    
    # Setup logging based on arguments
    global logger
    if args.log_file:
        # Setup file logging
        file_handler = logging.FileHandler(args.log_file)
        formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)
        
    if args.debug:
        logger.setLevel(logging.DEBUG)
        logging.getLogger().setLevel(logging.DEBUG)
        logger.debug("Debug logging enabled")
    
    logger.info(f"Starting file metadata extraction from: {args.directory}")
    logger.info(f"Output file: {args.output}")
    logger.info(f"Debug mode: {args.debug}")
    
    if not os.path.isdir(args.directory):
        logger.error(f"Directory not found: {args.directory}")
        print(f"Error: {args.directory} is not a valid directory")
        sys.exit(1)
    
    try:
        if args.csv_only:
            # CSV-only mode
            csv_file = args.output.replace('.xlsx', '.csv').replace('.xlsm', '.csv')
            logger.info("CSV-only mode enabled")
            process_files_to_excel(args.directory, csv_file, export_csv=True, litigant_name=args.litigant)
        else:
            # Normal mode (Excel + optional CSV)
            process_files_to_excel(args.directory, args.output, export_csv=args.csv, litigant_name=args.litigant)
            
        logger.info("Processing completed successfully")
        
    except PermissionError as e:
        logger.error(f"Permission error: {e}")
        print(f"\nERROR: Permission denied - {e}")
        print("\nTroubleshooting steps:")
        print("1. Close Excel or any program that might have the file open")
        print("2. Check file permissions and make sure directory is writable")
        print("3. Try running with --csv-only flag for CSV output only")
        print("4. Try specifying a different output directory with -o")
        sys.exit(1)
        
    except Exception as e:
        logger.error(f"Fatal error during processing: {e}")
        logger.error(f"Full traceback:\n{traceback.format_exc()}")
        print(f"\nERROR: {e}")
        if args.debug:
            print("\nCheck debug logs for detailed error information.")
        print("\nTry using --csv-only flag if Excel export is failing.")
        sys.exit(1)


if __name__ == "__main__":
    main()