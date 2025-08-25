#!/usr/bin/env python3

import os
import time
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import letter
from datetime import datetime, timedelta

def create_test_pdf(filename, num_pages, content_text):
    """Create a PDF with specified number of pages"""
    c = canvas.Canvas(filename, pagesize=letter)
    width, height = letter
    
    for page in range(num_pages):
        c.drawString(100, height - 100, f"Page {page + 1} of {num_pages}")
        c.drawString(100, height - 130, f"File: {filename}")
        c.drawString(100, height - 160, content_text)
        
        # Add some varied content to make different file sizes
        y_pos = height - 200
        for i in range(page * 5 + 10):  # Different amounts of text per page
            c.drawString(100, y_pos, f"This is line {i} of content to vary file size and make it realistic.")
            y_pos -= 20
            if y_pos < 100:
                break
                
        if page < num_pages - 1:
            c.showPage()
    
    c.save()

def main():
    # Create test PDFs with different page counts
    test_pdfs = [
        ("01SmallReport.pdf", 1, "Single page PDF document"),
        ("02MediumGuide.pdf", 5, "Medium-sized PDF with multiple pages"),
        ("03LargeManual.pdf", 15, "Large PDF manual with many pages"),
        ("04ShortDoc.pdf", 3, "Short document with few pages"),
        ("05ComprehensiveReport.pdf", 25, "Comprehensive report with many pages"),
        ("06QuickReference.pdf", 2, "Quick reference guide")
    ]
    
    print("Creating test PDF files...")
    for filename, pages, description in test_pdfs:
        create_test_pdf(filename, pages, description)
        print(f"Created {filename} with {pages} pages")
        
        # Modify file timestamps to create different creation dates
        # This simulates files created at different times
        days_ago = hash(filename) % 30  # Different dates within last 30 days
        timestamp = time.time() - (days_ago * 24 * 60 * 60)
        os.utime(filename, (timestamp, timestamp))

if __name__ == "__main__":
    try:
        main()
        print("\nAll test PDF files created successfully!")
    except ImportError:
        print("reportlab not found. Install with: pip install reportlab")
        print("Creating simple text files instead...")
        
        # Fallback: create simple text files with .pdf extension for testing
        test_files = [
            "01SmallReport.pdf",
            "02MediumGuide.pdf", 
            "03LargeManual.pdf",
            "04ShortDoc.pdf",
            "05ComprehensiveReport.pdf",
            "06QuickReference.pdf"
        ]
        
        for filename in test_files:
            with open(filename, 'w') as f:
                f.write(f"Test file: {filename}\n" * (hash(filename) % 100 + 10))
            print(f"Created text file: {filename}")