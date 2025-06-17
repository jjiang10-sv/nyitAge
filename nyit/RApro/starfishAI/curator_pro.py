
"""  
Curator Use Cases - Comprehensive Examples  
  
This file demonstrates the main use cases of the Curator library,  
showing how data scientists can use curator.LLM for various tasks.  
"""  
  
import os  
from typing import Dict, List, Literal  
from pydantic import BaseModel, Field  
from bespokelabs import curator  
#from bespokelabs.curator.types import Image  

backend_params = {
    "api_key" : ""
}
  
  
# Enable the Curator viewer for real-time visualization  
os.environ["CURATOR_VIEWER"] = "1"  
  
  
###########################################  
# 1. Bulk Inference with LLMs - Sentiment Analysis  
###########################################  
  
class Sentiment(BaseModel):  
    sentiment: Literal["positive", "negative", "neutral"] = Field(  
        description="Sentiment of the review")  
  
class SentimentAnalyzer(curator.LLM):  
    response_format = Sentiment  
      
    def prompt(self, product: Dict):  
        return f"Determine the sentiment of the product from the review: {product['review']}"  
      
    def parse(self, product: Dict, response: Sentiment):  
        return [{"name": product["name"], "sentiment": response.sentiment}]  
  
  
###########################################  
# 2. Structured Output Generation - Topic Generation  
###########################################  
  
class Topics(BaseModel):  
    topics_list: List[str] = Field(description="A list of topics.")  
  
class TopicGenerator(curator.LLM):  
    response_format = Topics  
      
    def prompt(self, subject):  
        return f"Return 3 topics related to {subject}"  
      
    def parse(self, input, response: Topics):  
        return [{"topic": t} for t in response.topics_list]  
  
  
###########################################  
# 3. Multi-Stage Data Generation Pipeline  
###########################################  
  
class Subject(BaseModel):  
    subject: str = Field(description="A subject")  
  
class Subjects(BaseModel):  
    subjects: List[Subject] = Field(description="A list of subjects")  
  
class QA(BaseModel):  
    question: str = Field(description="A question")  
    answer: str = Field(description="An answer")  
  
class QAs(BaseModel):  
    qas: List[QA] = Field(description="A list of QAs")  
  
class SubjectLLM(curator.LLM):  
    response_format = Subjects  
      
    def prompt(self, input: dict) -> str:  
        return "Generate a diverse list of 3 subjects. Keep it high-level (e.g. Math, Science)."  
      
    def parse(self, input: dict, response) -> dict:  
        return list(response.subjects)  
  
class SubsubjectLLM(curator.LLM):  
    response_format = Subjects  
      
    def prompt(self, input: dict) -> str:  
        return f"For the given subject {input['subject']}. Generate 3 diverse subsubjects. No explanation."  
      
    def parse(self, input: dict, response) -> dict:  
        return [{"subject": input["subject"], "subsubject": subsubject.subject} for subsubject in response.subjects]  
  
class QALLM(curator.LLM):  
    response_format = QAs  
      
    def prompt(self, input: dict) -> str:  
        return f"For the given subsubject {input['subsubject']}. Generate 3 diverse questions and answers. No explanation."  
      
    def parse(self, input: dict, response) -> dict:  
        return [  
            {  
                "subject": input["subject"],  
                "subsubject": input["subsubject"],  
                "question": qa.question,  
                "answer": qa.answer,  
            }  
            for qa in response.qas  
        ]  
  
  
###########################################  
# 4. Cost-Efficient Batch Processing - Poetry Generation  
###########################################  
  
class Poems(BaseModel):  
    poems_list: List[str] = Field(description="A list of poems.")  
  
class Poet(curator.LLM):  
    response_format = Poems  
      
    def prompt(self, input: dict) -> str:  
        return "Write two simple poems."  
      
    def parse(self, input: dict, response: Poems) -> dict:  
        return [{"poem": p} for p in response.poems_list]  
  
  
###########################################  
# 5. Reasoning and Thinking Trajectory Extraction  
###########################################  
  
class Reasoner(curator.LLM):  
    return_completions_object = True  
      
    def prompt(self, input):  
        return input["question"]  
      
    def parse(self, input, response):  
        """Parse the LLM response to extract reasoning and solution."""  
        content = response["content"]  
        thinking = ""  
        text = ""  
          
        for content_block in content:  
            if content_block["type"] == "thinking":  
                if "thinking" in content_block:  
                    thinking = content_block["thinking"]  
                else:  
                    thinking = content_block["text"]  
              
            elif content_block["type"] == "text":  
                text = content_block["text"]  
            elif content_block["type"] == "redacted_thinking":  
                print("Redacted thinking block! (notifying you for fun)")  
          
        input["model_thinking_trajectory"] = thinking  
        input["model_attempt"] = text  
        return input  
  
  
###########################################  
# 6. Multimodal Processing  
###########################################  
  
# class MultiModalLLM(curator.LLM):  
#     def prompt(self, input: dict):  
#         # Return text and image for multimodal processing  
#         return input["text"], Image(url=input["image_url"])  
      
#     def parse(self, input: dict, response) -> dict:  
#         return {"description": response}  
  
  
###########################################  
# 7. Text Analysis with Structured Output  
###########################################  
  
class Analysis(BaseModel):  
    summary: str = Field(description="A summary of the text")  
    key_points: List[str] = Field(description="Key points from the text")  
  
class TextAnalyzer(curator.LLM):  
    response_format = Analysis  
      
    def prompt(self, document):  
        return f"Analyze this document and extract key points: {document['text']}"  
      
    def parse(self, document, response):  
        return {  
            "title": document["title"],  
            "summary": response.summary,  
            "key_points": response.key_points  
        }  
  
  
def main():  
    """Run examples of all Curator use cases."""  
      
    print("1. Bulk Inference with LLMs - Sentiment Analysis")  
    analyzer = SentimentAnalyzer(model_name="gpt-4o-mini")  
    sentiment_dataset = [  
        {"name": "Curator", "review": "Already saved hours in one day of use."},  
        {"name": "Bespoke MiniCheck", "review": "Hallucination rates dropped by 90%."}  
    ]  
    sentiment_results = analyzer(sentiment_dataset)  
    print(sentiment_results.to_pandas())  
    print("\n")  
      
    print("2. Structured Output Generation - Topic Generation")  
    topic_generator = TopicGenerator(model_name="gpt-4o-mini")  
    topics = topic_generator("Mathematics")  
    print(topics.to_pandas())  
    print("\n")  
      
    print("3. Multi-Stage Data Generation Pipeline")  
    subject_prompter = SubjectLLM(model_name="gpt-4o-mini")  
    subsubject_prompter = SubsubjectLLM(model_name="gpt-4o-mini")  
    qa_prompter = QALLM(model_name="gpt-4o-mini")  
      
    subject_dataset = subject_prompter()  
    subsubject_dataset = subsubject_prompter(subject_dataset)  
    qa_dataset = qa_prompter(subsubject_dataset)  
    print(qa_dataset.to_pandas().head())  
    print("\n")  
      
    print("4. Cost-Efficient Batch Processing - Poetry Generation")  
    poet = Poet(  
        model_name="gpt-4o-mini",  
        batch=True,  # Enable batch processing to save 50% on token costs  
        backend_params={  
            "batch_check_interval": 60,  
            "delete_successful_batch_files": True  
        }  
    )  
    poems = poet(["Write poems about nature", "Write poems about technology"])  
    print(poems.to_pandas())  
    print("\n")  
      
    print("5. Multi-Provider Support Examples")  
    # OpenAI  
    openai_llm = curator.LLM(model_name="gpt-4o-mini")  
      
    # Anthropic via LiteLLM  
    anthropic_llm = curator.LLM(  
        model_name="anthropic/claude-3-5-sonnet-20240620",  
        backend="litellm",  
        backend_params={  
            "max_requests_per_minute": 100,  
            "max_tokens_per_minute": 1_000_000  
        }  
    )  
      
    # Local inference with vLLM  
    vllm_llm = curator.LLM(  
        model_name="Qwen/Qwen2.5-3B-Instruct",  
        backend="vllm",  
        backend_params={  
            "tensor_parallel_size": 1,  
            "gpu_memory_utilization": 0.7  
        }  
    )  
    print("Multi-provider LLMs initialized successfully")  
    print("\n")  
      
    print("6. Reasoning and Thinking Trajectory Extraction")  
    reasoner = Reasoner(  
        model_name="anthropic/claude-3-5-sonnet-20240620",  
        backend="litellm"  
    )  
    reasoning_results = reasoner([  
        {"question": "Solve this math problem: If a train travels at 60 mph, how long will it take to travel 150 miles?"}  
    ])  
    print(reasoning_results.to_pandas())  
    print("\n")  
      
    # print("7. Multimodal Processing")  
    # multimodal_llm = MultiModalLLM(model_name="gpt-4o")  
    # multimodal_results = multimodal_llm([  
    #     {"text": "Describe this image in detail:", "image_url": "https://example.com/image.jpg"}  
    # ])  
    # print(multimodal_results.to_pandas())  
    # print("\n")  
      
    print("8. Text Analysis with Structured Output")  
    analyzer = TextAnalyzer(model_name="gpt-4o-mini")  
    analysis_results = analyzer([  
        {"title": "Document 1", "text": "This is a sample document about AI..."},  
        {"title": "Document 2", "text": "Another document discussing machine learning..."}  
    ])  
    print(analysis_results.to_pandas())  
  
  
if __name__ == "__main__":  
    main()
# # curator_viewer = curator.Viewer()
# llm = curator.LLM(model_name="gpt-4o-mini", backend_params=backend_params)
# poem = llm("Write a poem about the importance of data in AI.")

# curator.view(poem)
# print(poem.to_pandas())
