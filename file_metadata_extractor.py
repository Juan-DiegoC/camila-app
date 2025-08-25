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
    from openpyxl import load_workbook, Workbook
    from openpyxl.utils import get_column_letter
except ImportError:
    print("openpyxl not found. Install with: pip install openpyxl==3.1.2")
    sys.exit(1)


def extract_number_from_filename(filename: str) -> int:
    """Extract the number from filenames like '01FileName', '02OtherFileName', etc."""
    match = re.match(r'^(\d+)', filename)
    return int(match.group(1)) if match else float('inf')


def get_pdf_pages(file_path: str) -> int:
    """Get number of pages in a PDF file."""
    try:
        with open(file_path, 'rb') as file:
            pdf_reader = PyPDF2.PdfReader(file)
            return len(pdf_reader.pages)
    except Exception:
        return 1


def get_file_size(file_path: str) -> int:
    """Get file size in bytes."""
    try:
        return os.path.getsize(file_path)
    except Exception:
        return 0


def count_files_in_directory(dir_path: str) -> int:
    """Count number of files in a directory."""
    try:
        return len([f for f in os.listdir(dir_path) if os.path.isfile(os.path.join(dir_path, f))])
    except Exception:
        return 0


def get_creation_date(file_path: str) -> datetime:
    """Get file creation date."""
    try:
        stat = os.stat(file_path)
        # Use birth time if available (macOS), otherwise use modification time
        creation_time = getattr(stat, 'st_birthtime', stat.st_mtime)
        return datetime.fromtimestamp(creation_time)
    except Exception:
        return datetime.now()


def get_ordered_files(directory: str) -> List[Tuple[str, str]]:
    """Get files and directories ordered by their numeric prefix."""
    items = []
    
    for item in os.listdir(directory):
        item_path = os.path.join(directory, item)
        items.append((item, item_path))
    
    # Sort by the numeric prefix extracted from filename
    items.sort(key=lambda x: extract_number_from_filename(x[0]))
    
    return items


def process_files_to_excel(directory: str, output_file: str = "metadata_output.xlsx"):
    """Process files and write metadata to Excel file."""
    
    # Get ordered files
    ordered_items = get_ordered_files(directory)
    
    # Create or load workbook
    try:
        wb = load_workbook(output_file)
        ws = wb.active
    except FileNotFoundError:
        wb = Workbook()
        ws = wb.active
        ws.title = "File Metadata"
    
    # Starting row (A12 as specified)
    start_row = 12
    
    for idx, (item_name, item_path) in enumerate(ordered_items):
        current_row = start_row + idx
        
        # Column A: File name
        ws[f'A{current_row}'] = item_name
        
        # Column B: Creation date (both B and next row B as specified)
        creation_date = get_creation_date(item_path)
        date_str = creation_date.strftime('%Y-%m-%d %H:%M:%S')
        ws[f'B{current_row}'] = date_str
        ws[f'B{current_row + 1}'] = date_str
        
        if os.path.isfile(item_path):
            # Column E: Number of pages (for PDF) or 1 for other files
            if item_path.lower().endswith('.pdf'):
                pages = get_pdf_pages(item_path)
            else:
                pages = 1
            ws[f'E{current_row}'] = pages
            
            # Column I: File size in bytes
            file_size = get_file_size(item_path)
            ws[f'I{current_row}'] = file_size
            
        elif os.path.isdir(item_path):
            # For directories: E column gets 1, I column gets file count
            ws[f'E{current_row}'] = 1
            file_count = count_files_in_directory(item_path)
            ws[f'I{current_row}'] = file_count
    
    # Save the workbook
    wb.save(output_file)
    print(f"Metadata extracted and saved to {output_file}")
    print(f"Processed {len(ordered_items)} items from {directory}")


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