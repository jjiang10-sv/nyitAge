# from PyPDF2 import PdfReader, PdfWriter
# from PyPDF2.generic import RectangleObject

# def fit_two_pages_into_one(input_path, output_path):
#     # Create PDF reader and get pages
#     reader = PdfReader(input_path)
#     if len(reader.pages) < 2:
#         raise ValueError("PDF must contain at least two pages")
    
#     page1 = reader.pages[0]
#     page2 = reader.pages[1]
    
#     # Create writer and add blank page
#     writer = PdfWriter()
#     original_width = page1.mediabox.width
#     original_height = page1.mediabox.height
    
#     # Create a new page with double width (for side-by-side)
#     new_page = writer.add_blank_page(
#         width=original_width * 2,  # Double width for two pages
#         height=original_height
#     )
    
#     # Scale and position pages
#     scale_factor = 0.5  # Scale to 50% of original size
    
#     # Add first page (left side)
#     new_page.merge_transformed_page(
#         page1,
#         (scale_factor, 0, 0, scale_factor, 0, 0),
#         expand=False
#     )
    
#     # Add second page (right side)
#     new_page.merge_transformed_page(
#         page2,
#         (scale_factor, 0, 0, scale_factor, original_width * scale_factor, 0),
#         expand=False
#     )
    
#     # Save output
#     with open(output_path, "wb") as out_file:
#         writer.write(out_file)

# # Usage example
# fit_two_pages_into_one("input.pdf", "output.pdf")