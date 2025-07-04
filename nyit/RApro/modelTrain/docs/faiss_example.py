import faiss
import numpy as np
from sentence_transformers import SentenceTransformer
import pickle
import time

class SemanticSearchEngine:
    """Example semantic search engine using FAISS"""
    
    def __init__(self, model_name='all-MiniLM-L6-v2'):
        self.model = SentenceTransformer(model_name)
        self.index = None
        self.documents = []
        self.embeddings = None
        
    def add_documents(self, documents):
        """Add documents to the search index"""
        print(f"Processing {len(documents)} documents...")
        
        # Generate embeddings
        self.documents.extend(documents)
        new_embeddings = self.model.encode(documents, show_progress_bar=True)
        
        if self.embeddings is None:
            self.embeddings = new_embeddings
        else:
            self.embeddings = np.vstack([self.embeddings, new_embeddings])
        
        # Build/rebuild index
        self._build_index()
        
    def _build_index(self):
        """Build FAISS index from embeddings"""
        dimension = self.embeddings.shape[1]
        
        if len(self.embeddings) < 1000:
            # Use exact search for small datasets
            self.index = faiss.IndexFlatL2(dimension)
            print("Using exact search (IndexFlatL2)")
        else:
            # Use approximate search for larger datasets
            nlist = min(100, len(self.embeddings) // 10)  # Number of clusters
            quantizer = faiss.IndexFlatL2(dimension)
            self.index = faiss.IndexIVFFlat(quantizer, dimension, nlist)
            
            # Train the index
            print("Training IVF index...")
            self.index.train(self.embeddings.astype('float32'))
            print("Using approximate search (IndexIVFFlat)")
        
        # Add vectors to index
        self.index.add(self.embeddings.astype('float32'))
        print(f"Index built with {self.index.ntotal} vectors")
    
    def search(self, query, top_k=5):
        """Search for similar documents"""
        if self.index is None:
            return []
        
        # Generate query embedding
        query_embedding = self.model.encode([query])
        
        # Search
        start_time = time.time()
        distances, indices = self.index.search(
            query_embedding.astype('float32'), top_k
        )
        search_time = time.time() - start_time
        
        # Format results
        results = []
        for i, (distance, idx) in enumerate(zip(distances[0], indices[0])):
            if idx != -1:  # Valid result
                results.append({
                    'rank': i + 1,
                    'document': self.documents[idx],
                    'distance': float(distance),
                    'similarity': 1 / (1 + distance)  # Convert distance to similarity
                })
        
        print(f"Search completed in {search_time:.4f} seconds")
        return results
    
    def save_index(self, filepath):
        """Save index and documents to disk"""
        # Save FAISS index
        faiss.write_index(self.index, f"{filepath}.faiss")
        
        # Save documents and embeddings
        with open(f"{filepath}_data.pkl", 'wb') as f:
            pickle.dump({
                'documents': self.documents,
                'embeddings': self.embeddings
            }, f)
        
        print(f"Index saved to {filepath}")
    
    def load_index(self, filepath):
        """Load index and documents from disk"""
        # Load FAISS index
        self.index = faiss.read_index(f"{filepath}.faiss")
        
        # Load documents and embeddings
        with open(f"{filepath}_data.pkl", 'rb') as f:
            data = pickle.load(f)
            self.documents = data['documents']
            self.embeddings = data['embeddings']
        
        print(f"Index loaded from {filepath}")
    
    def get_stats(self):
        """Get index statistics"""
        if self.index is None:
            return {"status": "No index built"}
        
        return {
            "total_documents": len(self.documents),
            "vector_dimension": self.embeddings.shape[1],
            "index_type": type(self.index).__name__,
            "is_trained": getattr(self.index, 'is_trained', True),
            "total_vectors": self.index.ntotal
        }

def demonstrate_faiss_indexing():
    """Demo different FAISS index types"""
    
    # Sample data
    dimension = 128
    n_vectors = 10000
    n_queries = 100
    
    print("=== FAISS Index Comparison ===\n")
    
    # Generate random data
    np.random.seed(42)
    database_vectors = np.random.random((n_vectors, dimension)).astype('float32')
    query_vectors = np.random.random((n_queries, dimension)).astype('float32')
    
    indexes = {
        "Flat L2 (Exact)": faiss.IndexFlatL2(dimension),
        "IVF100 (Approximate)": None,  # Will create below
        "LSH": faiss.IndexLSH(dimension, 64)
    }
    
    # Create IVF index
    quantizer = faiss.IndexFlatL2(dimension)
    ivf_index = faiss.IndexIVFFlat(quantizer, dimension, 100)
    ivf_index.train(database_vectors)
    indexes["IVF100 (Approximate)"] = ivf_index
    
    # Test each index type
    for name, index in indexes.items():
        print(f"--- {name} ---")
        
        # Add vectors
        start_time = time.time()
        index.add(database_vectors)
        add_time = time.time() - start_time
        
        # Search
        start_time = time.time()
        distances, indices = index.search(query_vectors[:10], k=5)
        search_time = time.time() - start_time
        
        print(f"  Add time: {add_time:.4f}s")
        print(f"  Search time: {search_time:.4f}s")
        print(f"  Memory usage: ~{index.ntotal * dimension * 4 / 1024 / 1024:.1f} MB")
        print(f"  Vectors stored: {index.ntotal}")
        print()

def main():
    """Main demo function"""
    print("🔍 FAISS Semantic Search Demo\n")
    
    # Sample documents
    documents = [
        "Machine learning is a subset of artificial intelligence",
        "Deep learning uses neural networks with multiple layers",
        "Natural language processing helps computers understand text",
        "Computer vision enables machines to interpret images",
        "Reinforcement learning trains agents through rewards",
        "Supervised learning uses labeled training data",
        "Unsupervised learning finds patterns in unlabeled data",
        "Transfer learning adapts pre-trained models",
        "Feature engineering improves model performance",
        "Cross-validation helps prevent overfitting"
    ]
    
    # Create search engine
    search_engine = SemanticSearchEngine()
    search_engine.add_documents(documents)
    
    # Display stats
    print("\n📊 Index Statistics:")
    stats = search_engine.get_stats()
    for key, value in stats.items():
        print(f"  {key}: {value}")
    
    # Perform searches
    queries = [
        "neural networks and deep learning",
        "training machine learning models",
        "working with images and vision"
    ]
    
    print("\n🔍 Search Results:")
    for query in queries:
        print(f"\nQuery: '{query}'")
        print("-" * 50)
        
        results = search_engine.search(query, top_k=3)
        for result in results:
            print(f"  {result['rank']}. {result['document']}")
            print(f"     Similarity: {result['similarity']:.3f}")
    
    # Demonstrate index comparison
    print("\n" + "="*60)
    demonstrate_faiss_indexing()

if __name__ == "__main__":
    main()