import bentoml  
bentoml.container.get_containerfile("synthetic_classifier:latest", output_path="./Dockerfile")