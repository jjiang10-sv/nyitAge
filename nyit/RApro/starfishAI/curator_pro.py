from bespokelabs import curator
backend_params = {
    "api_key" : ""
}
llm = curator.LLM(model_name="gpt-4o-mini")
poem = llm("Write a poem about the importance of data in AI.")
print(poem.to_pandas())

