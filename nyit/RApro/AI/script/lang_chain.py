from langchain_community.embeddings import HuggingFaceEmbeddings
from langchain_community.vectorstores import FAISS
from langchain_community.llms import Together
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.chains import RetrievalQA, ConversationalRetrievalChain
from langchain.prompts import PromptTemplate
from langchain.memory import ConversationBufferMemory
import os
from typing import List, Dict
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class AdvancedRAG:
    def __init__(
        self,
        model_name: str = "meta-llama/Llama-3.3-70B-Instruct-Turbo",
        embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2",
        chunk_size: int = 1000,
        chunk_overlap: int = 200,
        top_k: int = 3,
        temperature: float = 0.7,
    ):
        """
        Initialize the Advanced RAG system.
        
        Args:
            model_name: The LLM model to use
            embedding_model: The embedding model to use
            chunk_size: Size of text chunks
            chunk_overlap: Overlap between chunks
            top_k: Number of documents to retrieve
            temperature: LLM temperature
        """
        self.model_name = model_name
        self.embedding_model = embedding_model
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self.top_k = top_k
        self.temperature = temperature
        
        # Initialize components
        self._initialize_components()
        
    def _initialize_components(self):
        """Initialize all necessary components."""
        try:
            # Initialize text splitter
            self.text_splitter = RecursiveCharacterTextSplitter(
                chunk_size=self.chunk_size,
                chunk_overlap=self.chunk_overlap,
                length_function=len,
            )
            
            # Initialize embeddings
            self.embeddings = HuggingFaceEmbeddings(
                model_name=self.embedding_model,
                model_kwargs={'device': 'cpu'},
                encode_kwargs={'normalize_embeddings': True}
            )
            
            # Initialize LLM
            self.llm = Together(
                model=self.model_name,
                temperature=self.temperature,
                max_tokens=500,
            )
            
            # Initialize memory
            self.memory = ConversationBufferMemory(
                memory_key="chat_history",
                return_messages=True
            )
            
            logger.info("Successfully initialized all components")
            
        except Exception as e:
            logger.error(f"Error initializing components: {str(e)}")
            raise

    def process_documents(self, documents: Dict[str, str]) -> List[str]:
        """
        Process documents into chunks.
        
        Args:
            documents: Dictionary of document names and contents
            
        Returns:
            List of processed chunks
        """
        try:
            all_chunks = []
            for doc_name, content in documents.items():
                # Add document metadata
                chunks = self.text_splitter.split_text(content)
                chunks_with_metadata = [
                    f"[{doc_name}]\n{chunk}" for chunk in chunks
                ]
                all_chunks.extend(chunks_with_metadata)
            
            logger.info(f"Processed {len(documents)} documents into {len(all_chunks)} chunks")
            return all_chunks
            
        except Exception as e:
            logger.error(f"Error processing documents: {str(e)}")
            raise

    def create_vector_store(self, chunks: List[str]) -> FAISS:
        """
        Create vector store from chunks.
        
        Args:
            chunks: List of text chunks
            
        Returns:
            FAISS vector store
        """
        try:
            vector_store = FAISS.from_texts(
                chunks,
                self.embeddings,
                metadatas=[{"source": chunk.split("]")[0][1:]} for chunk in chunks]
            )
            logger.info("Successfully created vector store")
            return vector_store
            
        except Exception as e:
            logger.error(f"Error creating vector store: {str(e)}")
            raise

    def create_qa_chain(self, vector_store: FAISS) -> ConversationalRetrievalChain:
        """
        Create QA chain with memory.
        
        Args:
            vector_store: FAISS vector store
            
        Returns:
            ConversationalRetrievalChain
        """
        try:
            # Custom prompt template
            prompt_template = """
            You are an AI assistant that answers questions based on the provided context.
            
            Context: {context}
            
            Chat History: {chat_history}
            
            Question: {question}
            
            Instructions:
            1. Answer based ONLY on the provided context
            2. If the context doesn't contain enough information, say so
            3. Cite your sources using [document_name]
            4. Be concise but informative
            5. If the question is not related to the context, say so
            
            Answer:"""
            
            PROMPT = PromptTemplate(
                template=prompt_template,
                input_variables=["context", "chat_history", "question"]
            )
            
            # Create retriever
            retriever = vector_store.as_retriever(
                search_type="similarity",
                search_kwargs={"k": self.top_k}
            )
            
            # Create chain
            qa_chain = ConversationalRetrievalChain.from_llm(
                llm=self.llm,
                retriever=retriever,
                memory=self.memory,
                combine_docs_chain_kwargs={"prompt": PROMPT},
                return_source_documents=True,
                verbose=True
            )
            
            logger.info("Successfully created QA chain")
            return qa_chain
            
        except Exception as e:
            logger.error(f"Error creating QA chain: {str(e)}")
            raise

    def query(self, question: str, qa_chain: ConversationalRetrievalChain) -> Dict:
        """
        Query the RAG system.
        
        Args:
            question: User question
            qa_chain: QA chain
            
        Returns:
            Dictionary containing answer and sources
        """
        try:
            # Get response
            response = qa_chain({"question": question})
            
            # Extract sources
            sources = [doc.metadata["source"] for doc in response["source_documents"]]
            
            return {
                "answer": response["answer"],
                "sources": list(set(sources)),  # Remove duplicates
                "chat_history": response["chat_history"]
            }
            
        except Exception as e:
            logger.error(f"Error querying RAG system: {str(e)}")
            raise

def main():
    # Example usage
    documents = {
        "octopus_facts": "Octopuses have three hearts and blue blood. Two hearts pump blood to the gills, while the third pumps blood to the rest of the body. Their blood is blue because it contains copper-based hemocyanin instead of iron-based hemoglobin.",
        "honey_facts": "Honey never spoils. Archaeologists have found pots of honey in ancient Egyptian tombs that are over 3,000 years old and still perfectly edible. This is because honey has natural antibacterial properties and very low water content.",
        "space_facts": "A day on Venus is longer than its year. Venus takes 243 Earth days to rotate once on its axis, but only 225 Earth days to orbit the Sun. This means a Venusian day is longer than a Venusian year."
    }
    
    # Initialize RAG system
    rag = AdvancedRAG()
    
    # Process documents
    chunks = rag.process_documents(documents)
    
    # Create vector store
    vector_store = rag.create_vector_store(chunks)
    
    # Create QA chain
    qa_chain = rag.create_qa_chain(vector_store)
    
    # Example questions
    questions = [
        "What's special about octopus blood?",
        "How long can honey last?",
        "What's interesting about Venus?"
    ]
    
    # Query the system
    for question in questions:
        print(f"\nQuestion: {question}")
        response = rag.query(question, qa_chain)
        print(f"Answer: {response['answer']}")
        print(f"Sources: {response['sources']}")
        print("-" * 50)

if __name__ == "__main__":
    main()