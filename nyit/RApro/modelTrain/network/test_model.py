import requests


# Convert to list for JSON serialization
data = [{
      "dst_host_count": 255,
      "dst_host_serror_rate": 0.00,
      "dst_host_srv_serror_rate": 0.00,
      "dst_host_same_src_port_rate": 0.11,
      "srv_count": 9,
      "dst_host_diff_srv_rate": 0.05,
      "logged_in": 1,
      "dst_host_srv_count": 255,
      "dst_bytes": 5450,
      "src_bytes": 181
}]

# Send POST request to BentoML API
response = requests.post(
    #"http://localhost:3000/predict",  # Update if your endpoint is different
    "https://my-synthetic-classifier-67630700.mt-guc1.bentoml.ai/predict",
    json={"flow": data}
)

print("Status code:", response.status_code)
print("Predictions:", response.json())