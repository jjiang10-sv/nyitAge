# Web Traffic Prediction System

A machine learning system for predicting web traffic based on Apache log data using LSTM (Long Short-Term Memory) neural networks.

## Overview

This system analyzes historical Apache access logs to predict future traffic patterns. It can forecast traffic for the next few minutes or hours, enabling:

- **Auto-scaling**: Automatically scale infrastructure based on predicted load
- **Proactive alerting**: Get warnings before traffic spikes occur
- **Resource planning**: Plan capacity upgrades based on trends
- **Cost optimization**: Scale down during predicted low-traffic periods
- **Anomaly detection**: Identify unusual traffic patterns

## Model Architecture

The system uses an LSTM neural network with the following architecture:

```
Input: Last 60 time periods (default: 5 hours with 5-minute windows)
       ↓
LSTM Layer (128 units) + Dropout (0.2)
       ↓
LSTM Layer (64 units) + Dropout (0.2)
       ↓
LSTM Layer (32 units) + Dropout (0.2)
       ↓
Dense Layer (64 units)
       ↓
Output: Next 12 time periods (default: 1 hour)
```

## Files

- **[`traffic_predictor.py`](traffic_predictor.py)** - Main training script and TrafficPredictor class
- **[`demo_prediction.py`](demo_prediction.py)** - Demo showing how to use the trained model
- **[`requirements.txt`](requirements.txt)** - Python dependencies
- **[`dataset/nasa.txt`](dataset/nasa.txt)** - Large NASA Apache log dataset for training
- **[`dataset/easy.txt`](dataset/easy.txt)** - Small sample dataset

## Installation

```bash
# Install dependencies
pip install -r requirements.txt
```

Required packages:
- pandas >= 2.0.0
- numpy >= 1.24.0
- matplotlib >= 3.7.0
- scikit-learn >= 1.3.0
- tensorflow >= 2.13.0
- joblib >= 1.3.0

## Quick Start

### 1. Train the Model

```bash
python traffic_predictor.py
```

This will:
- Parse the Apache log file (`dataset/nasa.txt`)
- Extract timestamps and count requests per time window
- Create training sequences
- Train the LSTM model
- Save the trained model and scaler
- Generate training history plots

**Training Configuration** (editable in `main()` function):
```python
TIME_WINDOW = '5min'         # Time aggregation window
LOOKBACK_STEPS = 60          # Number of past periods to use
PREDICTION_HORIZON = 12      # Number of future periods to predict
SAMPLE_RATE = 0.1            # Use 10% of data (adjust based on memory)
EPOCHS = 30                  # Training epochs
BATCH_SIZE = 32              # Batch size
```

### 2. Make Predictions

```bash
python demo_prediction.py
```

This demonstrates:
- Loading the trained model
- Making predictions based on recent traffic
- Real-time monitoring scenarios
- Batch predictions for multiple scenarios

## Usage Examples

### Basic Prediction

```python
from traffic_predictor import TrafficPredictor
import numpy as np

# Load trained model
predictor = TrafficPredictor()
predictor.load_model()

# Recent traffic observations (last 60 periods)
recent_traffic = np.array([100, 105, 98, 110, ...])  # 60 values

# Predict next 12 periods
prediction = predictor.predict(recent_traffic)

print(f"Predicted traffic: {prediction}")
```

### Real-time Monitoring

```python
# Define thresholds
NORMAL = 150
HIGH = 200

# Make prediction
prediction = predictor.predict(recent_traffic)

# Check each predicted period
for i, traffic in enumerate(prediction):
    if traffic > HIGH:
        print(f"🚨 CRITICAL: Period {i+1}: {traffic:.0f} requests")
        # Trigger scaling/alerting
    elif traffic > NORMAL:
        print(f"⚠️  WARNING: Period {i+1}: {traffic:.0f} requests")
```

### Custom Training

```python
from traffic_predictor import TrafficPredictor

# Custom configuration
predictor = TrafficPredictor(
    time_window='10min',      # 10-minute windows
    lookback_steps=48,        # Look back 8 hours
    prediction_horizon=6      # Predict next hour
)

# Train with custom parameters
predictor.train(
    log_file='path/to/access.log',
    sample_rate=1.0,          # Use all data
    epochs=50,
    batch_size=64,
    validation_split=0.2
)

# Save model
predictor.save_model('my_custom_model.keras')
```

## Apache Log Format

The system parses standard Apache Combined Log Format:

```
199.0.2.27 - - [28/Jul/1995:13:32:20 -0400] "GET /images/NASA-logosmall.gif HTTP/1.0" 200 786
```

Extracted fields:
- IP address
- Timestamp: `[28/Jul/1995:13:32:20 -0400]`
- HTTP method and path
- Status code
- Response size

## Configuration Options

### Time Windows

Common configurations:

| Window | Lookback | Horizon | Use Case |
|--------|----------|---------|----------|
| 1min | 60 | 15 | High-frequency monitoring |
| 5min | 60 | 12 | Standard web apps (recommended) |
| 10min | 48 | 6 | Lower frequency sites |
| 1hour | 24 | 4 | Daily pattern prediction |

### Model Parameters

Adjust in [`traffic_predictor.py`](traffic_predictor.py):

```python
# Data preparation
sample_rate = 0.1     # Use 10% of data (increase if you have enough memory)

# Model architecture (in build_model())
LSTM(128, ...)        # Increase for more complex patterns
Dropout(0.2)          # Increase to reduce overfitting

# Training
epochs = 30           # Increase for better convergence
batch_size = 32       # Adjust based on GPU memory
```

## Output Files

After training, the following files are created:

- **`traffic_model.keras`** - Trained LSTM model
- **`traffic_scaler.pkl`** - MinMax scaler for normalization
- **`traffic_model_metadata.pkl`** - Model configuration
- **`best_traffic_model.keras`** - Best model from training (lowest val loss)
- **`training_history.png`** - Loss and MAE plots

## Performance

Expected metrics (depends on data):
- Training MAE: 5-15 requests per period
- Validation MAE: 10-20 requests per period
- Prediction time: < 50ms per forecast

## Integration Examples

### Kubernetes Auto-scaling

```python
# Get prediction
prediction = predictor.predict(recent_traffic)
predicted_avg = prediction.mean()

# Scale based on prediction
if predicted_avg > 200:
    replicas = 5
elif predicted_avg > 150:
    replicas = 3
else:
    replicas = 1

# Update deployment
kubectl.scale_deployment('web-app', replicas=replicas)
```

### Alerting System

```python
import smtplib

prediction = predictor.predict(recent_traffic)

# Check for spike
if max(prediction) > SPIKE_THRESHOLD:
    send_alert(
        to='ops-team@company.com',
        subject='Traffic spike predicted',
        body=f'Expected peak: {max(prediction):.0f} requests'
    )
```

### Load Balancer Configuration

```python
# Predict traffic
prediction = predictor.predict(recent_traffic)

# Adjust backend pool
if prediction.mean() > HIGH_TRAFFIC:
    loadbalancer.add_backend_servers(count=3)
elif prediction.mean() < LOW_TRAFFIC:
    loadbalancer.remove_backend_servers(count=2)
```

## Troubleshooting

### Not Enough Data

```
ValueError: Not enough data points. Found only 16 timestamps.
```

**Solution**: Use a larger dataset or reduce `lookback_steps` and `prediction_horizon`.

### Memory Error

```
MemoryError: Unable to allocate array
```

**Solution**: Reduce `sample_rate` in training (e.g., `sample_rate=0.05` for 5% of data).

### Poor Predictions

**Symptoms**: High MAE, predictions don't follow patterns

**Solutions**:
1. Increase `sample_rate` to use more training data
2. Increase `epochs` for better convergence
3. Adjust `time_window` to match your traffic patterns
4. Try different `lookback_steps` values
5. Add more features (day of week, time of day, etc.)

## Advanced Features

### Multi-variate Prediction

Extend the model to include additional features:

```python
def prepare_sequences_multivariate(self, data, features):
    """
    Prepare sequences with multiple features
    features: dict with 'traffic', 'day_of_week', 'hour', etc.
    """
    # Stack features
    X = np.column_stack([features[key] for key in features])
    
    # Reshape for LSTM: [samples, timesteps, features]
    return X
```

### Transfer Learning

Fine-tune on new data:

```python
# Load pre-trained model
predictor.load_model('traffic_model.keras')

# Fine-tune on new data
predictor.model.compile(optimizer='adam', loss='mse', lr=0.0001)
predictor.train(new_log_file, epochs=10)
```

## Data Pipeline

Typical production pipeline:

```
Apache Logs → Log Parser → Time Series DB → Feature Engineering → Model → Predictions
                                                                               ↓
                                                                    Monitoring/Alerting
                                                                               ↓
                                                                    Auto-scaling System
```

## Citation

If you use this system in your research, please cite:

```bibtex
@software{traffic_predictor_2024,
  title = {Web Traffic Prediction using LSTM},
  author = {Your Name},
  year = {2024},
  url = {https://github.com/yourusername/traffic-predictor}
}
```

## License

MIT License - feel free to use and modify for your needs.

## Support

For issues or questions:
- Check the [Troubleshooting](#troubleshooting) section
- Review the [demo_prediction.py](demo_prediction.py) examples
- Open an issue on GitHub

## Future Enhancements

- [ ] Add Prophet model as alternative to LSTM
- [ ] Implement online learning for continuous improvement
- [ ] Add seasonal decomposition for better long-term predictions
- [ ] Support for multiple log formats (Nginx, IIS, etc.)
- [ ] Web dashboard for real-time monitoring
- [ ] Integration with Prometheus/Grafana
- [ ] Anomaly detection module
- [ ] Multi-step ahead forecasting with uncertainty estimation