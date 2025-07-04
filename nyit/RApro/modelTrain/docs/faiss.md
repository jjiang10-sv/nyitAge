# FAISS: Facebook AI Similarity Search

**FAISS** (Facebook AI Similarity Search) is a powerful library developed by Facebook AI Research for efficient similarity search and clustering of dense vectors. It's particularly popular in machine learning for vector databases and semantic search applications.

## **What FAISS Does**

FAISS specializes in:
- **Fast similarity search** in large collections of vectors
- **Nearest neighbor search** (finding similar embeddings)
- **Vector clustering** and indexing
- **Memory-efficient** storage of high-dimensional vectors

## **How FAISS Indexing Works**

### **1. Basic Concept**
```python
import faiss
import numpy as np

# Create sample vectors (1000 vectors of 128 dimensions)
vectors = np.random.random((1000, 128)).astype('float32')

# Build index
dimension = 128
index = faiss.IndexFlatL2(dimension)  # L2 distance
index.add(vectors)  # Add vectors to index

# Search for similar vectors
query = np.random.random((1, 128)).astype('float32')
distances, indices = index.search(query, k=5)  # Find 5 nearest
```

### **2. Index Types**

#### **Flat Indexes (Exact Search)**
```python
# L2 (Euclidean) distance
index = faiss.IndexFlatL2(dimension)

# Inner product (cosine similarity when normalized)
index = faiss.IndexFlatIP(dimension)
```

#### **IVF (Inverted File) - Approximate Search**
```python
# Faster but approximate
quantizer = faiss.IndexFlatL2(dimension)
index = faiss.IndexIVFFlat(quantizer, dimension, nlist=100)

# Train the index first
index.train(vectors)
index.add(vectors)
```

#### **LSH (Locality Sensitive Hashing)**
```python
# For binary vectors or very high dimensions
index = faiss.IndexLSH(dimension, nbits=128)
```

## **Real-World Example: Semantic Search**
Now let me clean up the example file:
Ran tool
# FAISS: Facebook AI Similarity Search Library

**FAISS** is a library developed by Facebook AI Research for efficient similarity search and clustering of dense vectors. It's widely used in machine learning for vector databases, recommendation systems, and semantic search.

## **What FAISS Does**

### **Core Capabilities:**
- **Fast similarity search** in large vector collections (millions to billions)
- **Nearest neighbor search** with various distance metrics
- **Memory-efficient** indexing of high-dimensional vectors
- **GPU acceleration** for massive datasets
- **Approximate search** algorithms for speed vs accuracy trade-offs

## **How FAISS Indexing Works**

### **1. Basic Indexing Process**

```python
import faiss
import numpy as np

# 1. Create vectors (embeddings)
vectors = np.random.random((10000, 128)).astype('float32')

# 2. Choose and build index
dimension = 128
index = faiss.IndexFlatL2(dimension)  # L2 distance

# 3. Add vectors to index
index.add(vectors)

# 4. Search for similar vectors
query = np.random.random((1, 128)).astype('float32')
distances, indices = index.search(query, k=5)  # Find 5 nearest
```

### **2. Index Types & Performance**

#### **Exact Search Indexes**
```python
# Flat L2 - Exact Euclidean distance
index = faiss.IndexFlatL2(dimension)

# Flat IP - Exact inner product (cosine similarity)
index = faiss.IndexFlatIP(dimension)
```
- **Pros:** 100% accurate results
- **Cons:** O(n) search time, doesn't scale to millions of vectors

#### **Approximate Search Indexes**
```python
# IVF (Inverted File) - Partitions vectors into clusters
quantizer = faiss.IndexFlatL2(dimension)
index = faiss.IndexIVFFlat(quantizer, dimension, nlist=100)

# Must train before adding vectors
index.train(training_vectors)
index.add(vectors)

# Control speed vs accuracy
index.nprobe = 10  # Search 10 clusters (default: 1)
```
- **Pros:** Much faster search O(log n)
- **Cons:** Approximate results (99%+ accuracy typical)

#### **Compressed Indexes**
```python
# Product Quantization - Compresses vectors
index = faiss.IndexIVFPQ(quantizer, dimension, nlist=100, m=8, nbits=8)
```
- **Pros:** 10-100x memory reduction
- **Cons:** Some accuracy loss

### **3. Real-World Usage Patterns**

#### **Semantic Search Engine**
```python
from sentence_transformers import SentenceTransformer

# 1. Generate embeddings
model = SentenceTransformer('all-MiniLM-L6-v2')
documents = ["doc1 text", "doc2 text", ...]
embeddings = model.encode(documents)

# 2. Build FAISS index
index = faiss.IndexFlatL2(embeddings.shape[1])
index.add(embeddings.astype('float32'))

# 3. Search
query_embedding = model.encode(["search query"])
distances, indices = index.search(query_embedding, k=10)

# 4. Get results
results = [documents[i] for i in indices[0]]
```

#### **Recommendation System**
```python
# User/item embeddings from collaborative filtering
user_embeddings = np.array(...)  # Shape: (n_users, embedding_dim)
item_embeddings = np.array(...)  # Shape: (n_items, embedding_dim)

# Build item index
item_index = faiss.IndexFlatIP(embedding_dim)  # Cosine similarity
item_index.add(item_embeddings)

# Find similar items for a user
user_vector = user_embeddings[user_id:user_id+1]
_, similar_items = item_index.search(user_vector, k=20)
```

## **Key FAISS Features for Large-Scale Systems**

### **1. Persistence**
```python
# Save index to disk
faiss.write_index(index, "my_index.faiss")

# Load index from disk
index = faiss.read_index("my_index.faiss")
```

### **2. GPU Acceleration**
```python
# Move index to GPU
gpu_index = faiss.index_cpu_to_gpu(res, 0, index)  # GPU 0

# Or build directly on GPU
index = faiss.GpuIndexFlatL2(res, dimension)
```

### **3. Memory Mapping (for huge datasets)**
```python
# Memory-mapped index for datasets larger than RAM
index = faiss.read_index("huge_index.faiss", faiss.IO_FLAG_MMAP)
```

## **Performance Comparison**

| Index Type | Search Time | Memory Usage | Accuracy | Use Case |
|------------|-------------|--------------|----------|----------|
| `IndexFlatL2` | O(n) | High | 100% | Small datasets (<100K) |
| `IndexIVFFlat` | O(log n) | High | 95-99% | Medium datasets (100K-10M) |
| `IndexIVFPQ` | O(log n) | Low | 90-95% | Large datasets (10M+) |
| `IndexLSH` | O(1) | Medium | 85-95% | Very large, binary data |

## **Common Use Cases**

### **1. Document/Text Search**
- Convert documents to embeddings using BERT/Sentence-BERT
- Build FAISS index for fast semantic search
- Used by: search engines, Q&A systems

### **2. Image Similarity**
- Use CNN features as embeddings
- Find visually similar images
- Used by: photo apps, e-commerce

### **3. Recommendation Systems**
- User/item embeddings from collaborative filtering
- Fast nearest neighbor lookup
- Used by: Netflix, Spotify, Amazon

### **4. Vector Databases**
- Foundation for systems like Pinecone, Weaviate
- Real-time similarity search APIs
- Used by: RAG systems, chatbots

## **Integration with Other Libraries**

### **With LangChain**
```python
from langchain.vectorstores import FAISS
from langchain.embeddings import OpenAIEmbeddings

# Create vector store
embeddings = OpenAIEmbeddings()
vector_store = FAISS.from_texts(texts, embeddings)

# Search
results = vector_store.similarity_search("query", k=5)
```

### **With Sentence Transformers**
```python
from sentence_transformers import SentenceTransformer
import faiss

model = SentenceTransformer('all-MiniLM-L6-v2')
embeddings = model.encode(documents)

index = faiss.IndexFlatL2(embeddings.shape[1])
index.add(embeddings)
```

## **Best Practices**

### **1. Choosing the Right Index**
- **Small datasets (< 100K):** Use `IndexFlatL2` or `IndexFlatIP`
- **Medium datasets (100K - 10M):** Use `IndexIVFFlat`
- **Large datasets (10M+):** Use `IndexIVFPQ` with compression
- **Real-time updates:** Use indices that support `add()` operations

### **2. Optimization Tips**
```python
# Normalize vectors for cosine similarity
faiss.normalize_L2(vectors)

# Tune IVF parameters
index.nprobe = 32  # Search more clusters for better accuracy

# Use appropriate data types
vectors = vectors.astype('float32')  # FAISS requires float32
```

### **3. Memory Management**
```python
# For large datasets, use memory mapping
index = faiss.read_index("index.faiss", faiss.IO_FLAG_MMAP)

# Or use compressed indices
index = faiss.IndexIVFPQ(quantizer, d, nlist, m, nbits)
```

## **Real Performance Numbers**

For a typical semantic search setup:
- **Dataset:** 1M documents, 384-dim embeddings
- **IndexFlatL2:** ~100ms search time, 1.5GB memory
- **IndexIVFFlat:** ~5ms search time, 1.5GB memory, 98% accuracy
- **IndexIVFPQ:** ~3ms search time, 150MB memory, 95% accuracy

**FAISS is essential for any application dealing with vector similarity search at scale!** It's the backbone of modern semantic search, recommendation systems, and RAG (Retrieval Augmented Generation) applications.

# SentenceTransformer Library Guide

**SentenceTransformer** is a Python library that provides an easy way to compute dense vector representations (embeddings) for sentences, paragraphs, and entire documents. It's built on top of PyTorch and Transformers.

## What is SentenceTransformer?

SentenceTransformer converts text into high-dimensional vectors that capture semantic meaning, enabling:
- **Semantic similarity search** - Find similar text based on meaning, not just keywords
- **Text clustering** - Group similar documents together
- **Information retrieval** - Build search engines that understand context
- **Question answering** - Match questions to relevant content

## How It Works

```python
from sentence_transformers import SentenceTransformer

# Load a pre-trained model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Encode sentences into embeddings
sentences = [
    "The cat sits on the mat",
    "A feline rests on a carpet", 
    "Dogs are playing in the park"
]

embeddings = model.encode(sentences)
print(f"Shape: {embeddings.shape}")  # (3, 384) - 3 sentences, 384 dimensions
```

## Popular Models

| Model | Dimensions | Speed | Quality | Use Case |
|-------|------------|-------|---------|----------|
| `all-MiniLM-L6-v2` | 384 | Fast | Good | General purpose |
| `all-mpnet-base-v2` | 768 | Medium | Best | High accuracy needed |
| `paraphrase-MiniLM-L6-v2` | 384 | Fast | Good | Paraphrase detection |
| `multi-qa-MiniLM-L6-cos-v1` | 384 | Fast | Good | Question-answering |

## Your Implementation

In your code, SentenceTransformer is used to create embeddings for text chunks:

```python:nyit/RApro/modelTrain/memvid/memvid/index.py
def __init__(self, config: Optional[Dict[str, Any]] = None):
    self.config = config or get_default_config()
    self.embedding_model = SentenceTransformer(self.config["embedding"]["model"])
    self.dimension = self.config["embedding"]["dimension"]
```

## Core Usage Patterns

### 1. **Basic Encoding**

```python
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('all-MiniLM-L6-v2')

# Single sentence
text = "Machine learning is fascinating"
embedding = model.encode(text)
print(f"Embedding shape: {embedding.shape}")  # (384,)

# Multiple sentences
texts = ["Hello world", "How are you?", "Python is great"]
embeddings = model.encode(texts)
print(f"Batch shape: {embeddings.shape}")  # (3, 384)
```

### 2. **Semantic Similarity**

```python
from sentence_transformers import SentenceTransformer, util

model = SentenceTransformer('all-MiniLM-L6-v2')

# Encode sentences
sentences = [
    "I love programming",
    "Coding is my passion", 
    "I hate vegetables"
]

embeddings = model.encode(sentences)

# Compute similarity
similarity = util.cos_sim(embeddings[0], embeddings[1])
print(f"Similarity: {similarity.item():.3f}")  # High similarity (~0.7)

similarity = util.cos_sim(embeddings[0], embeddings[2])  
print(f"Similarity: {similarity.item():.3f}")  # Low similarity (~0.1)
```

### 3. **Your Robust Implementation**

Your code includes excellent error handling for embedding generation:

```python:nyit/RApro/modelTrain/memvid/memvid/index.py
def _generate_embeddings(self, chunks: List[str], show_progress: bool) -> np.ndarray:
    """Generate embeddings with error handling and batch processing"""
    
    # Try full batch first
    try:
        logger.info(f"Generating embeddings for {len(chunks)} chunks (full batch)")
        embeddings = self.embedding_model.encode(
            chunks,
            show_progress_bar=show_progress,
            batch_size=32,
            convert_to_numpy=True,
            normalize_embeddings=True  # Important for FAISS
        )
        return np.array(embeddings).astype('float32')
    
    except Exception as e:
        logger.warning(f"Full batch embedding failed: {e}. Trying batch processing...")
        return self._generate_embeddings_batched(chunks, show_progress)
```

## Advanced Features

### 1. **Batch Processing with Parameters**

```python
model = SentenceTransformer('all-MiniLM-L6-v2')

embeddings = model.encode(
    sentences,
    batch_size=32,              # Process 32 sentences at once
    show_progress_bar=True,     # Show progress
    convert_to_numpy=True,      # Return numpy array
    normalize_embeddings=True,  # L2 normalize (good for cosine similarity)
    device='cuda'               # Use GPU if available
)
```

### 2. **Custom Models and Fine-tuning**

```python
# Load from local path or Hugging Face Hub
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

# Fine-tune on your data
from sentence_transformers import InputExample, losses

# Create training examples
train_examples = [
    InputExample(texts=['Anchor text', 'Positive text'], label=1.0),
    InputExample(texts=['Anchor text', 'Negative text'], label=0.0)
]

# Fine-tune
train_dataloader = DataLoader(train_examples, shuffle=True, batch_size=16)
train_loss = losses.CosineSimilarityLoss(model)
model.fit([(train_dataloader, train_loss)], epochs=1, warmup_steps=100)
```

### 3. **Multilingual Support**

```python
# Multilingual model
model = SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')

texts = [
    "Hello, how are you?",           # English
    "Hola, ¿cómo estás?",          # Spanish  
    "Bonjour, comment allez-vous?", # French
    "こんにちは、元気ですか？"            # Japanese
]

embeddings = model.encode(texts)
# All embeddings in same vector space!
```

## Integration with FAISS (Your Use Case)

Your implementation combines SentenceTransformer with FAISS perfectly:

```python
# 1. Generate embeddings with SentenceTransformer
embeddings = self.embedding_model.encode(chunks)

# 2. Store in FAISS for fast similarity search
self.index.add_with_ids(embeddings, chunk_ids)

# 3. Search with query
def search(self, query: str, top_k: int = 5):
    # Convert query to embedding
    query_embedding = self.embedding_model.encode([query])
    
    # Search in FAISS
    distances, indices = self.index.search(query_embedding, top_k)
    return results
```

## Performance Considerations

### 1. **Model Selection**

```python
# Fast but lower quality
model = SentenceTransformer('all-MiniLM-L6-v2')  # 384 dim, ~80MB

# Slower but higher quality  
model = SentenceTransformer('all-mpnet-base-v2')  # 768 dim, ~420MB

# Specialized for questions
model = SentenceTransformer('multi-qa-MiniLM-L6-cos-v1')  # 384 dim
```

### 2. **Batch Processing**

```python
# Good batch sizes
embeddings = model.encode(
    large_text_list,
    batch_size=32,  # Adjust based on GPU memory
    show_progress_bar=True,
    convert_to_numpy=True
)
```

### 3. **Memory Management**

```python
# For very large datasets, process in chunks
def encode_large_dataset(texts, model, chunk_size=10000):
    all_embeddings = []
    
    for i in range(0, len(texts), chunk_size):
        chunk = texts[i:i + chunk_size]
        embeddings = model.encode(chunk, convert_to_numpy=True)
        all_embeddings.append(embeddings)
    
    return np.vstack(all_embeddings)
```

## Your System Architecture

Your implementation creates a powerful semantic search system:

1. **Text Processing**: EPUB/PDF → Text chunks
2. **Embedding**: SentenceTransformer → Dense vectors  
3. **Indexing**: FAISS → Fast similarity search
4. **Storage**: QR codes → Video frames
5. **Retrieval**: Query → Similar chunks → Video timestamps

This enables semantic search over video-encoded documents! Users can search by meaning rather than exact keywords, making your system much more powerful than traditional text search.

The combination of SentenceTransformer + FAISS is a proven pattern for building production-ready semantic search systems.