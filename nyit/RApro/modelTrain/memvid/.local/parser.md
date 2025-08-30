For running OCR on a server, you have several integration options. I'll show you how to modify the code to support both local Tesseract and cloud-based OCR services for better server deployment.

⁠ python:nyit/RApro/modelTrain/memvid/memvid/encoder.py
    def add_pdf(self, pdf_path: str, chunk_size: int = DEFAULT_CHUNK_SIZE, overlap: int = DEFAULT_OVERLAP,
                extract_image_text: bool = False, extract_table_text: bool = False, 
                ocr_backend: str = "tesseract"):
        """
        Extract text, image text (OCR), and table text from PDF and add as chunks

        Args:
            pdf_path: Path to PDF file
            chunk_size: Target chunk size
            overlap: Overlap between chunks
            extract_image_text: Whether to extract text from images using OCR
            extract_table_text: Whether to extract text from tables
            ocr_backend: OCR backend ('tesseract', 'google', 'aws', 'azure')
        """
        # Check dependencies
        missing_deps = []
        try:
            import fitz  # PyMuPDF
        except ImportError:
            missing_deps.append("PyMuPDF")

        if extract_table_text:
            try:
                import pdfplumber
            except ImportError:
                missing_deps.append("pdfplumber")

        if extract_image_text:
            if ocr_backend == "tesseract":
                try:
                    import pytesseract
                    from PIL import Image
                    import io
                except ImportError:
                    missing_deps.append("pytesseract pillow")
            elif ocr_backend == "google":
                try:
                    from google.cloud import vision
                    from PIL import Image
                    import io
                except ImportError:
                    missing_deps.append("google-cloud-vision pillow")
            elif ocr_backend == "aws":
                try:
                    import boto3
                    from PIL import Image
                    import io
                except ImportError:
                    missing_deps.append("boto3 pillow")
            elif ocr_backend == "azure":
                try:
                    from azure.cognitiveservices.vision.computervision import ComputerVisionClient
                    from msrest.authentication import CognitiveServicesCredentials
                    from PIL import Image
                    import io
                except ImportError:
                    missing_deps.append("azure-cognitiveservices-vision-computervision pillow")

        if missing_deps:
            raise ImportError(f"Missing dependencies: {', '.join(missing_deps)}. "
                            f"Install with: pip install {' '.join(missing_deps)}")

        if not Path(pdf_path).exists():
            raise FileNotFoundError(f"PDF file not found: {pdf_path}")

        text_content = []
        image_text_count = 0
        table_count = 0

        try:
            # Open PDF with PyMuPDF for text and images
            pdf_document = fitz.open(pdf_path)
            num_pages = len(pdf_document)

            logger.info(f"Extracting content from {num_pages} pages of {Path(pdf_path).name}")

            # Also open with pdfplumber for table extraction if needed
            if extract_table_text:
                import pdfplumber
                plumber_pdf = pdfplumber.open(pdf_path)

            for page_num in range(num_pages):
                page_text_parts = []
                
                # === BASE TEXT EXTRACTION with PyMuPDF ===
                fitz_page = pdf_document[page_num]
                base_text = fitz_page.get_text("text")
                if base_text.strip():
                    page_text_parts.append(base_text)

                # === IMAGE TEXT EXTRACTION with OCR ===
                if extract_image_text:
                    image_list = fitz_page.get_images(full=True)
                    
                    for img_index, img in enumerate(image_list):
                        try:
                            # Extract image data
                            xref = img[0]
                            base_image = pdf_document.extract_image(xref)
                            image_bytes = base_image["image"]
                            
                            # Perform OCR based on backend
                            ocr_text = self._extract_text_from_image(image_bytes, ocr_backend)
                            
                            if ocr_text and ocr_text.strip():
                                image_text_count += 1
                                image_text_header = f"\n--- Text from Image {image_text_count} on page {page_num + 1} ---\n"
                                image_text_footer = f"\n--- End Image Text {image_text_count} ---\n"
                                
                                full_image_text = image_text_header + ocr_text.strip() + image_text_footer
                                page_text_parts.append(full_image_text)
                                
                        except Exception as e:
                            logger.warning(f"Failed to extract text from image {img_index + 1} on page {page_num + 1}: {e}")
                            continue

                # === TABLE TEXT EXTRACTION with pdfplumber ===
                if extract_table_text:
                    plumber_page = plumber_pdf.pages[page_num]
                    tables = plumber_page.extract_tables()
                    
                    for table_index, table in enumerate(tables):
                        if table:  # Ensure table is not empty
                            table_count += 1
                            
                            # Convert table to readable text format
                            table_header = f"\n--- Table {table_count} on page {page_num + 1} ---\n"
                            
                            # Format table as text with proper alignment
                            table_text_rows = []
                            for row_index, row in enumerate(table):
                                # Clean and join cells, handling None values
                                clean_row = [str(cell).strip() if cell else "" for cell in row]
                                
                                # Treat first row as header if it looks like one
                                if row_index == 0:
                                    table_text_rows.append(" | ".join(clean_row))
                                    table_text_rows.append("-" * len(" | ".join(clean_row)))  # Separator line
                                else:
                                    table_text_rows.append(" | ".join(clean_row))
                            
                            table_text = "\n".join(table_text_rows)
                            table_footer = f"\n--- End Table {table_count} ---\n"
                            
                            full_table_text = table_header + table_text + table_footer
                            page_text_parts.append(full_table_text)

                # Combine all text parts for this page
                if page_text_parts:
                    full_page_text = "\n".join(page_text_parts)
                    text_content.append(full_page_text)

            # Close documents
            pdf_document.close()
            if extract_table_text:
                plumber_pdf.close()

            # Combine all text content and add to chunks
            full_text = "\n\n".join(text_content)

            if full_text.strip():
                self.add_text(full_text, chunk_size, overlap)
                
                # Log extraction summary
                content_summary = []
                content_summary.append(f"{len(full_text)} characters")
                if image_text_count > 0:
                    content_summary.append(f"text from {image_text_count} images via {ocr_backend}")
                if table_count > 0:
                    content_summary.append(f"{table_count} tables")
                
                logger.info(f"Added PDF content from {Path(pdf_path).name}: {', '.join(content_summary)}")
            else:
                logger.warning(f"No content extracted from PDF: {pdf_path}")

            # Return extraction statistics
            return {
                'text_length': len(full_text),
                'images_with_text': image_text_count,
                'tables_extracted': table_count,
                'pages_processed': num_pages,
                'chunks_created': len(chunk_text(full_text, chunk_size, overlap)),
                'ocr_backend': ocr_backend
            }

        except Exception as e:
            logger.error(f"Error processing PDF {pdf_path}: {e}")
            raise

    def _extract_text_from_image(self, image_bytes: bytes, backend: str = "tesseract") -> str:
        """
        Extract text from image using specified OCR backend
        
        Args:
            image_bytes: Raw image bytes
            backend: OCR backend ('tesseract', 'google', 'aws', 'azure')
            
        Returns:
            Extracted text
        """
        from PIL import Image
        import io
        
        try:
            if backend == "tesseract":
                return self._ocr_tesseract(image_bytes)
            elif backend == "google":
                return self._ocr_google_vision(image_bytes)
            elif backend == "aws":
                return self._ocr_aws_textract(image_bytes)
            elif backend == "azure":
                return self._ocr_azure_computer_vision(image_bytes)
            else:
                raise ValueError(f"Unsupported OCR backend: {backend}")
                
        except Exception as e:
            logger.error(f"OCR failed with {backend}: {e}")
            return ""

    def _ocr_tesseract(self, image_bytes: bytes) -> str:
        """Extract text using Tesseract OCR"""
        import pytesseract
        from PIL import Image
        import io
        
        image = Image.open(io.BytesIO(image_bytes))
        return pytesseract.image_to_string(image, lang='eng')

    def _ocr_google_vision(self, image_bytes: bytes) -> str:
        """Extract text using Google Cloud Vision API"""
        from google.cloud import vision
        
        client = vision.ImageAnnotatorClient()
        image = vision.Image(content=image_bytes)
        
        response = client.text_detection(image=image)
        texts = response.text_annotations
        
        if texts:
            return texts[0].description
        return ""

    def _ocr_aws_textract(self, image_bytes: bytes) -> str:
        """Extract text using AWS Textract"""
        import boto3
        
        client = boto3.client('textract')
        
        response = client.detect_document_text(
            Document={'Bytes': image_bytes}
        )
        
        text_blocks = []
        for block in response['Blocks']:
            if block['BlockType'] == 'LINE':
                text_blocks.append(block['Text'])
        
        return '\n'.join(text_blocks)

    def _ocr_azure_computer_vision(self, image_bytes: bytes) -> str:
        """Extract text using Azure Computer Vision"""
        from azure.cognitiveservices.vision.computervision import ComputerVisionClient
        from msrest.authentication import CognitiveServicesCredentials
        import os
        import time
        
        # Initialize client
        subscription_key = os.environ.get('AZURE_COMPUTER_VISION_KEY')
        endpoint = os.environ.get('AZURE_COMPUTER_VISION_ENDPOINT')
        
        if not subscription_key or not endpoint:
            raise ValueError("Azure credentials not found in environment variables")
        
        client = ComputerVisionClient(endpoint, CognitiveServicesCredentials(subscription_key))
        
        # Call API
        read_response = client.read_in_stream(io.BytesIO(image_bytes), raw=True)
        
        # Get operation location
        read_operation_location = read_response.headers["Operation-Location"]
        operation_id = read_operation_location.split("/")[-1]
        
        # Wait for result
        while True:
            read_result = client.get_read_result(operation_id)
            if read_result.status not in ['notStarted', 'running']:
                break
            time.sleep(1)
        
        # Extract text
        text_blocks = []
        if read_result.status == 'succeeded':
            for text_result in read_result.analyze_result.read_results:
                for line in text_result.lines:
                    text_blocks.append(line.text)
        
        return '\n'.join(text_blocks)
 ⁠

## *Server Deployment Options*

### *1. Docker Container with Tesseract (Recommended for servers)*

Create a ⁠ Dockerfile ⁠:
⁠ dockerfile
FROM python:3.9-slim

# Install system dependencies for Tesseract
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-eng \
    libtesseract-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . /app
WORKDIR /app

CMD ["python", "your_app.py"]
 ⁠

### *2. Environment Variables for Cloud OCR*

⁠ bash
# For Google Cloud Vision
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"

# For AWS Textract
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_DEFAULT_REGION="us-east-1"

# For Azure Computer Vision
export AZURE_COMPUTER_VISION_KEY="your-key"
export AZURE_COMPUTER_VISION_ENDPOINT="https://your-region.cognitiveservices.azure.com/"
 ⁠

### *3. Installation Commands for Different Backends*

⁠ bash
# Local Tesseract (requires system installation)
sudo apt-get install tesseract-ocr tesseract-ocr-eng  # Ubuntu/Debian
brew install tesseract  # macOS
pip install pytesseract pillow

# Cloud services
pip install google-cloud-vision  # Google
pip install boto3  # AWS
pip install azure-cognitiveservices-vision-computervision  # Azure
 ⁠

### *4. Usage Examples*

⁠ python
# Local OCR (good for development/small servers)
encoder.add_pdf("document.pdf", extract_image_text=True, ocr_backend="tesseract")

# Cloud OCR (better for production servers)
encoder.add_pdf("document.pdf", extract_image_text=True, ocr_backend="google")
encoder.add_pdf("document.pdf", extract_image_text=True, ocr_backend="aws")
encoder.add_pdf("document.pdf", extract_image_text=True, ocr_backend="azure")

# Combined extraction
encoder.add_pdf("document.pdf", 
                extract_image_text=True, 
                extract_table_text=True, 
                ocr_backend="google")
 ⁠

*For production servers, I recommend using cloud-based OCR services as they:*
•⁠  ⁠Don't require system-level dependencies
•⁠  ⁠Have better accuracy
•⁠  ⁠Handle more image formats
•⁠  ⁠Scale automatically
•⁠  ⁠Provide better error handling

Would you like me to help you set up any specific OCR backend for your server environment?