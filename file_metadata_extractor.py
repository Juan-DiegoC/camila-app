#!/usr/bin/env python3

import os
import sys
import argparse
from datetime import datetime
from pathlib import Path
import re
from typing import List, Tuple

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


def is_valid_pdf(file_path: str) -> bool:
    """Check if file is a valid PDF by reading first few bytes."""
    try:
        with open(file_path, 'rb') as file:
            header = file.read(5)
            return header == b'%PDF-'
    except Exception:
        return False

def get_pdf_pages_pypdf2(file_path: str) -> int:
    """Get PDF pages using PyPDF2."""
    try:
        with open(file_path, 'rb') as file:
            pdf_reader = PyPDF2.PdfReader(file)
            return len(pdf_reader.pages)
    except Exception as e:
        print(f"PyPDF2 error for {file_path}: {str(e)[:100]}...")
        return None

def get_pdf_pages_pymupdf(file_path: str) -> int:
    """Get PDF pages using PyMuPDF (fallback)."""
    if not PYMUPDF_AVAILABLE:
        return None
    try:
        doc = fitz.open(file_path)
        page_count = len(doc)
        doc.close()
        return page_count
    except Exception as e:
        print(f"PyMuPDF error for {file_path}: {str(e)[:100]}...")
        return None

def get_pdf_pages_estimate(file_path: str) -> int:
    """Estimate PDF pages by searching for page objects in raw content."""
    try:
        with open(file_path, 'rb') as file:
            content = file.read()
            # Count occurrences of page object patterns
            page_patterns = [b'/Type /Page', b'/Type/Page', b'endobj']
            max_count = 0
            for pattern in page_patterns:
                count = content.count(pattern)
                if pattern == b'endobj':
                    count = max(1, count // 10)  # Rough estimate
                max_count = max(max_count, count)
            return max(1, min(max_count, 1000))  # Cap at reasonable number
    except Exception:
        return 1

def get_pdf_pages(file_path: str) -> int:
    """Get number of pages in a PDF file with multiple fallback methods."""
    # First check if it's a valid PDF
    if not is_valid_pdf(file_path):
        print(f"Warning: {file_path} is not a valid PDF file")
        return 1
    
    # Try PyPDF2 first
    pages = get_pdf_pages_pypdf2(file_path)
    if pages is not None:
        return pages
    
    # Try PyMuPDF as fallback
    pages = get_pdf_pages_pymupdf(file_path)
    if pages is not None:
        print(f"Used PyMuPDF fallback for {file_path}")
        return pages
    
    # Last resort: estimate based on file content
    pages = get_pdf_pages_estimate(file_path)
    print(f"Used estimation method for {file_path}: {pages} pages")
    return pages


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

def add_headers_and_formatting(ws):
    """Add headers and formatting to match the template."""
    header_font, header_alignment, data_font, data_alignment, thin_border = setup_excel_formatting()
    
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

def process_files_to_excel(directory: str, output_file: str = "metadata_output.xlsx"):
    """Process files and write metadata to Excel file matching the template format."""
    
    # Get ordered files
    ordered_items = get_ordered_files(directory)
    
    # Create new workbook (always start fresh to match template)
    wb = Workbook()
    ws = wb.active
    ws.title = "Indice Electrónico"
    
    # Add headers and formatting
    add_headers_and_formatting(ws)
    
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
            if item_path.lower().endswith('.pdf'):
                pages = get_pdf_pages(item_path)
            else:
                pages = 1
            ws[f'E{current_row}'].value = pages
            
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
            ws[f'H{current_row}'].value = "DIRECTORIO"
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
    
    # Save the workbook
    wb.save(output_file)
    print(f"Metadata extracted and saved to {output_file}")
    print(f"Processed {len(ordered_items)} items from {directory}")
    print(f"Format matches 'indice de ejemplo.xlsm' template")


def main():
    parser = argparse.ArgumentParser(description='Extract file metadata and write to Excel')
    parser.add_argument('--directory', '-d', 
                       default=os.getcwd(),
                       help='Directory to process (default: current directory)')
    parser.add_argument('--output', '-o',
                       default='metadata_output.xlsx',
                       help='Output Excel file (default: metadata_output.xlsx)')
    
    args = parser.parse_args()
    
    if not os.path.isdir(args.directory):
        print(f"Error: {args.directory} is not a valid directory")
        sys.exit(1)
    
    try:
        process_files_to_excel(args.directory, args.output)
    except Exception as e:
        print(f"Error processing files: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()