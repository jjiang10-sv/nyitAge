import random
import pandas as pd

from starfish import StructuredLLM, data_factory

from starfish.common.env_loader import load_env_file ## Load environment variables from .env file
load_env_file()

import datetime
# Create a StructuredLLM instance for generating network intrusion detection datapoints
intrusion_llm = StructuredLLM(
    model_name="openai/gpt-4o-mini",
    prompt="""You are an expert network security analyst and intrusion detection system monitor. 
    
    Generate realistic network traffic datapoints that could represent either normal traffic or various types of network intrusions. 
    
    Consider the following attack types:
    - Normal traffic (0)
    - DoS (Denial of Service) attacks (1) 
    - R2L (Remote to Local) attacks (2)
    - U2R (User to Root) attacks (3)
    - Probe attacks (4)
    
    For each datapoint, generate realistic values for network traffic features that would be characteristic of the specified traffic type.
    
    Generate a datapoint for traffic type: {{traffic_type}}""",
    
    output_schema=[
        {"name": "dst_host_count", "type": "int", "description": "Number of connections to the same destination host"},
        {"name": "dst_host_serror_rate", "type": "float", "description": "Percentage of connections to the same destination host that have SYN errors"},
        {"name": "dst_host_srv_serror_rate", "type": "float", "description": "Percentage of connections to the same destination host and service that have SYN errors"},
        {"name": "dst_host_same_src_port_rate", "type": "float", "description": "Percentage of connections to the same destination host from the same source port"},
        {"name": "srv_count", "type": "int", "description": "Number of connections to the same service"},
        {"name": "dst_host_diff_srv_rate", "type": "float", "description": "Percentage of connections to the same destination host that use different services"},
        {"name": "logged_in", "type": "int", "description": "1 if successfully logged in, 0 otherwise"},
        {"name": "dst_host_srv_count", "type": "int", "description": "Number of connections to the same destination host and service"},
        {"name": "dst_bytes", "type": "int", "description": "Number of data bytes sent from destination to source"},
        {"name": "src_bytes", "type": "int", "description": "Number of data bytes sent from source to destination"},
        {"name": "labels", "type": "int", "description": "Attack type: 0=normal, 1=DoS, 2=R2L, 3=U2R, 4=probe"}
    ],
    model_kwargs={"temperature": 0.3},
)

@data_factory(max_concurrency=5)
async def generate_intrusion_data(traffic_type: str):
    #print(f"Generating {traffic_type} traffic datapoint at {datetime.now()}")
    response = await intrusion_llm.run(traffic_type=traffic_type)
    return response.data

def synthetic_data_gen(num:int):
    # Generate datapoints for different traffic types
    base_traffic_types = ["normal", "DoS attack", "R2L attack", "U2R attack", "probe attack"]

    # Create a list of 100 randomly selected traffic types
    traffic_types = random.choices(base_traffic_types, k=num)

    # Optional: You can also control the distribution if needed
    # For example, to have more normal traffic and fewer rare attacks:
    # traffic_types = random.choices(
    #     base_traffic_types, 
    #     weights=[0.6, 0.2, 0.1, 0.05, 0.05],  # 60% normal, 20% DoS, etc.
    #     k=num
    # )
    generated_intrusion_dataset = generate_intrusion_data.run(traffic_type=traffic_types)
    return generated_intrusion_dataset

if __name__ == "__main__":
    data = synthetic_data_gen(100)
    print(data[0])
    df = pd.DataFrame(data)
    df.to_csv("synthetic_intrusion_data_1.csv", index=False)