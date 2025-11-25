

from docling.document_converter import DocumentConverter
from pathlib import Path
converter = DocumentConverter()
input_path = Path("/Users/john/Downloads/GDPR_Governance_1.pdf")
result = converter.convert(input_path)
# To export to Markdown
markdown_output_path = Path("output.md")
with markdown_output_path.open("w") as fp:
    fp.write(result.document.export_to_markdown())
#pandoc to docx
