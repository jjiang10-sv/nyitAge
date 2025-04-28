from bespokelabs import curator
backend_params = {
    "api_key" : ""
}
# curator_viewer = curator.Viewer()
llm = curator.LLM(model_name="gpt-4o-mini", backend_params=backend_params)
poem = llm("Write a poem about the importance of data in AI.")

curator.view(poem)
print(poem.to_pandas())
