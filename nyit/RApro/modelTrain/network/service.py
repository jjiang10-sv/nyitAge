import bentoml
import numpy as np
from typing import List
import joblib
import pandas as pd
from scipy.stats import zscore
#from hubble import HubbleClient#

@bentoml.service(
    image=bentoml.images.Image(python_version="3.11")
        .python_packages("scikit-learn", "numpy"),
    resources={"cpu": "1", "memory": "512Mi"}
)
class IntrusionDetection:
    # Load the saved model from Model Store
    model_ref = bentoml.models.BentoModel("intrusion_detector:latest")
    
    def __init__(self):
        # Load the actual model during service initialization
        self.model = joblib.load(self.model_ref.path_of("intrusion_detector_model.pkl"))

        self.expected_columns = ['dst_host_count', 'dst_host_serror_rate', 'dst_host_srv_serror_rate',
       'dst_host_same_src_port_rate', 'srv_count', 'dst_host_diff_srv_rate',
       'logged_in', 'dst_host_srv_count', 'dst_bytes', 'src_bytes']
    
    
    # def predict(self, features: List[List[float]]) -> List[str]:
    #     # Convert input to numpy array
    #     X = np.array(features)
        
    #     # Make predictions
    #     predictions = self.model.predict(X)
        
    #     # Convert to class names
    #     return [self.classes[p] for p in predictions]
    
    # @bentoml.api
    # def predict_proba(self, features: List[List[float]]) -> List[List[float]]:
    #     # Get prediction probabilities
    #     X = np.array(features)
    #     probabilities = self.model.predict_proba(X)
    #     return probabilities.tolist()
    

    def extract_features(self,flow):
        protocol_type = flow.get('l4', {}).get('protocol', 'other')
        service = flow.get('l7', {}).get('http', {}).get('method', str(flow.get('l4', {}).get('tcp', {}).get('destination_port', 'unknown')))
        flag = ','.join(flow.get('l4', {}).get('tcp', {}).get('flags', [])) if 'tcp' in flow.get('l4', {}) else 'none'
        src_bytes = 0
        dst_bytes = 0
        return {
            'protocol_type': protocol_type,
            'service': service,
            'flag': flag,
            'src_bytes': src_bytes,
            'dst_bytes': dst_bytes,
        }



    def preprocess_and_predict(self,flow):
        # features = self.extract_features(flow)
        # df = pd.DataFrame([features])
        # df_encoded = pd.get_dummies(df, columns=['protocol_type', 'service', 'flag'])
        # for col in self.expected_columns:
        #     if col not in df_encoded:
        #         df_encoded[col] = 0
        # df_encoded = df_encoded[self.expected_columns]

        traffic = pd.DataFrame([flow]).values
        traffic = np.nan_to_num(traffic).astype(np.float64)
        traffic = zscore(traffic, axis=0, ddof=1)
        pred =self.model.predict(traffic)[0]
        class_map = {0: 'Normal', 1: 'DoS', 2: 'R2L', 3: 'U2R', 4: 'Probe'}
        return class_map.get(pred, 'Unknown')

    @bentoml.api(batchable=True, max_batch_size=32)
    def predict(self, flow):
        #flow = request.json
        label = self.preprocess_and_predict(flow[0])
        return [{'prediction': label}]

# # Real-time flow consumer (runs in background)
# def real_time_consumer():
#     client = HubbleClient('localhost:4245')  # Adjust as needed
#     for flow in client.observe():
#         label = preprocess_and_predict(flow)
#         if label != 'Normal':
#             print(f"ALERT: {label} detected! Flow: {flow}")



