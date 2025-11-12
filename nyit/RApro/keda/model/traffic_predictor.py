"""
Web Traffic Prediction Model using LSTM
Predicts future traffic based on historical Apache log data
"""

import re
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
import matplotlib.pyplot as plt
from sklearn.preprocessing import MinMaxScaler
from sklearn.model_selection import train_test_split
import tensorflow as tf
import keras
from keras.models import Sequential
from keras.layers import LSTM, Dense, Dropout
from keras.callbacks import EarlyStopping, ModelCheckpoint
import joblib
import os

class TrafficPredictor:
    def __init__(self, time_window='1min', lookback_steps=60, prediction_horizon=10):
        """
        Initialize the traffic predictor
        
        Args:
            time_window: Time aggregation window ('1min', '5min', '1H', etc.)
            lookback_steps: Number of past time steps to use for prediction
            prediction_horizon: Number of future steps to predict
        """
        self.time_window = time_window
        self.lookback_steps = lookback_steps
        self.prediction_horizon = prediction_horizon
        self.scaler = MinMaxScaler(feature_range=(0, 1))
        self.model = None
        
    def parse_apache_log(self, log_file, sample_rate=1.0):
        """
        Parse Apache log file and extract timestamps
        
        Args:
            log_file: Path to the Apache log file
            sample_rate: Fraction of lines to sample (1.0 = all lines, 0.1 = 10%)
        """
        # Apache log pattern
        log_pattern = r'(\S+) - - \[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4})\]'
        
        timestamps = []
        line_count = 0
        
        print(f"Parsing log file: {log_file}")
        
        with open(log_file, 'r', encoding='utf-8', errors='ignore') as f:
            for line in f:
                line_count += 1
                
                # Sample lines based on sample_rate
                if sample_rate < 1.0 and np.random.random() > sample_rate:
                    continue
                
                match = re.search(log_pattern, line)
                if match:
                    timestamp_str = match.group(2)
                    try:
                        # Parse timestamp: 28/Jul/1995:13:32:20 -0400
                        dt = datetime.strptime(timestamp_str, '%d/%b/%Y:%H:%M:%S %z')
                        timestamps.append(dt)
                    except Exception as e:
                        continue
                
                # Progress indicator
                if line_count % 100000 == 0:
                    print(f"  Processed {line_count:,} lines, found {len(timestamps):,} valid entries")
        
        print(f"Total lines processed: {line_count:,}")
        print(f"Valid timestamps extracted: {len(timestamps):,}")
        
        return timestamps
    
    def create_time_series(self, timestamps):
        """
        Convert timestamps to time series with traffic counts per time window
        """
        # Create DataFrame with timestamps
        df = pd.DataFrame({'timestamp': timestamps})
        df['count'] = 1
        
        # Convert to timezone-naive for easier processing
        df['timestamp'] = pd.to_datetime(df['timestamp']).dt.tz_localize(None)
        
        # Set timestamp as index
        df.set_index('timestamp', inplace=True)
        
        # Resample to specified time window and count requests
        traffic_series = df['count'].resample(self.time_window).sum()
        
        # Fill missing periods with 0
        traffic_series = traffic_series.fillna(0)
        
        print(f"\nTime series created:")
        print(f"  Date range: {traffic_series.index.min()} to {traffic_series.index.max()}")
        print(f"  Total time periods: {len(traffic_series)}")
        print(f"  Average traffic per {self.time_window}: {traffic_series.mean():.2f} requests")
        print(f"  Max traffic: {traffic_series.max():.0f} requests")
        
        return traffic_series
    
    def prepare_sequences(self, data):
        """
        Prepare sequences for LSTM training
        Creates overlapping sequences of (lookback_steps) for prediction
        """
        X, y = [], []
        
        for i in range(len(data) - self.lookback_steps - self.prediction_horizon + 1):
            # Input sequence
            X.append(data[i:i + self.lookback_steps])
            # Target sequence (next prediction_horizon steps)
            y.append(data[i + self.lookback_steps:i + self.lookback_steps + self.prediction_horizon])
        
        return np.array(X), np.array(y)
    
    def build_model(self):
        """
        Build LSTM model for traffic prediction
        """
        model = Sequential([
            LSTM(128, activation='relu', return_sequences=True, 
                 input_shape=(self.lookback_steps, 1)),
            Dropout(0.2),
            LSTM(64, activation='relu', return_sequences=True),
            Dropout(0.2),
            LSTM(32, activation='relu'),
            Dropout(0.2),
            Dense(64, activation='relu'),
            Dense(self.prediction_horizon)
        ])
        
        model.compile(optimizer='adam', loss='mse', metrics=['mae'])
        
        print("\nModel Architecture:")
        model.summary()
        
        return model
    
    def train(self, log_file, sample_rate=1.0, epochs=50, batch_size=32, validation_split=0.2):
        """
        Train the traffic prediction model
        """
        print("="*60)
        print("TRAFFIC PREDICTION MODEL TRAINING")
        print("="*60)
        
        # Parse logs
        timestamps = self.parse_apache_log(log_file, sample_rate)
        
        if len(timestamps) < 100:
            raise ValueError(f"Not enough data points. Found only {len(timestamps)} timestamps.")
        
        # Create time series
        traffic_series = self.create_time_series(timestamps)
        
        # Normalize data
        traffic_data = np.array(traffic_series).reshape(-1, 1)
        traffic_scaled = self.scaler.fit_transform(traffic_data)
        
        # Prepare sequences
        print(f"\nPreparing sequences with lookback={self.lookback_steps}, horizon={self.prediction_horizon}")
        X, y = self.prepare_sequences(traffic_scaled)
        
        print(f"  Training sequences: {len(X)}")
        print(f"  Input shape: {X.shape}")
        print(f"  Output shape: {y.shape}")
        
        if len(X) == 0:
            raise ValueError("Not enough data to create training sequences. Try reducing lookback_steps or increase sample_rate.")
        
        # Reshape for LSTM [samples, timesteps, features]
        X = X.reshape((X.shape[0], X.shape[1], 1))
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=validation_split, shuffle=False
        )
        
        print(f"\nTraining set: {len(X_train)} sequences")
        print(f"Test set: {len(X_test)} sequences")
        
        # Build model
        self.model = self.build_model()
        
        # Callbacks
        early_stop = EarlyStopping(monitor='val_loss', patience=10, restore_best_weights=True)
        checkpoint = ModelCheckpoint(
            'best_traffic_model.keras',
            monitor='val_loss',
            save_best_only=True,
            verbose=1
        )
        
        # Train model
        print("\n" + "="*60)
        print("Training...")
        print("="*60)
        
        history = self.model.fit(
            X_train, y_train,
            epochs=epochs,
            batch_size=batch_size,
            validation_data=(X_test, y_test),
            callbacks=[early_stop, checkpoint],
            verbose="auto"
        )
        
        # Evaluate
        print("\n" + "="*60)
        print("Evaluation")
        print("="*60)
        
        train_loss, train_mae = self.model.evaluate(X_train, y_train, verbose="auto")
        test_loss, test_mae = self.model.evaluate(X_test, y_test, verbose="auto")
        
        print(f"Training Loss: {train_loss:.4f}, MAE: {train_mae:.4f}")
        print(f"Test Loss: {test_loss:.4f}, MAE: {test_mae:.4f}")
        
        # Plot training history
        self.plot_training_history(history)
        
        return history
    
    def plot_training_history(self, history):
        """
        Plot training and validation loss
        """
        plt.figure(figsize=(12, 4))
        
        plt.subplot(1, 2, 1)
        plt.plot(history.history['loss'], label='Training Loss')
        plt.plot(history.history['val_loss'], label='Validation Loss')
        plt.title('Model Loss')
        plt.xlabel('Epoch')
        plt.ylabel('Loss')
        plt.legend()
        plt.grid(True)
        
        plt.subplot(1, 2, 2)
        plt.plot(history.history['mae'], label='Training MAE')
        plt.plot(history.history['val_mae'], label='Validation MAE')
        plt.title('Model MAE')
        plt.xlabel('Epoch')
        plt.ylabel('MAE')
        plt.legend()
        plt.grid(True)
        
        plt.tight_layout()
        plt.savefig('training_history.png')
        print("\nTraining history plot saved as 'training_history.png'")
        plt.close()
    
    def predict(self, recent_data):
        """
        Predict future traffic based on recent observations
        
        Args:
            recent_data: Array of recent traffic counts (should have lookback_steps elements)
        
        Returns:
            Predicted traffic for next prediction_horizon time steps
        """
        if self.model is None:
            raise ValueError("Model not trained. Call train() first.")
        
        # Normalize
        recent_scaled = self.scaler.transform(recent_data.reshape(-1, 1))
        
        # Reshape for LSTM
        X = recent_scaled[-self.lookback_steps:].reshape(1, self.lookback_steps, 1)
        
        # Predict
        prediction_scaled = self.model.predict(X, verbose="auto")
        
        # Inverse transform
        prediction = self.scaler.inverse_transform(prediction_scaled)
        
        return prediction[0]
    
    def save_model(self, model_path='traffic_model.keras', scaler_path='traffic_scaler.pkl'):
        """
        Save trained model and scaler
        """
        if self.model is None:
            raise ValueError("No model to save. Train the model first.")
        
        self.model.save(model_path)
        joblib.dump(self.scaler, scaler_path)
        
        # Save metadata
        metadata = {
            'time_window': self.time_window,
            'lookback_steps': self.lookback_steps,
            'prediction_horizon': self.prediction_horizon
        }
        joblib.dump(metadata, 'traffic_model_metadata.pkl')
        
        print(f"\nModel saved to: {model_path}")
        print(f"Scaler saved to: {scaler_path}")
        print(f"Metadata saved to: traffic_model_metadata.pkl")
    
    def load_model(self, model_path='traffic_model.keras', scaler_path='traffic_scaler.pkl'):
        """
        Load trained model and scaler
        """
        self.model = keras.models.load_model(model_path)
        self.scaler = joblib.load(scaler_path)
        
        # Load metadata
        metadata = joblib.load('traffic_model_metadata.pkl')
        self.time_window = metadata['time_window']
        self.lookback_steps = metadata['lookback_steps']
        self.prediction_horizon = metadata['prediction_horizon']
        
        print(f"Model loaded from: {model_path}")
        print(f"Configuration: window={self.time_window}, lookback={self.lookback_steps}, horizon={self.prediction_horizon}")


def main():
    """
    Main training script
    """
    # Configuration
    LOG_FILE = 'bentoml/nyitAge/nyit/RApro/keda/model/dataset/nasa.txt'
    TIME_WINDOW = '5min'  # Aggregate traffic every 5 minutes
    LOOKBACK_STEPS = 60   # Use last 60 time periods (5 hours if 5min windows)
    PREDICTION_HORIZON = 12  # Predict next 12 periods (1 hour if 5min windows)
    SAMPLE_RATE = 0.1     # Use 10% of data for faster training (adjust as needed)
    EPOCHS = 30
    BATCH_SIZE = 32
    
    # Initialize predictor
    predictor = TrafficPredictor(
        time_window=TIME_WINDOW,
        lookback_steps=LOOKBACK_STEPS,
        prediction_horizon=PREDICTION_HORIZON
    )
    
    # Train model
    try:
        history = predictor.train(
            log_file=LOG_FILE,
            sample_rate=SAMPLE_RATE,
            epochs=EPOCHS,
            batch_size=BATCH_SIZE,
            validation_split=0.2
        )
        
        # Save model
        predictor.save_model()
        
        print("\n" + "="*60)
        print("TRAINING COMPLETE!")
        print("="*60)
        print("\nYou can now use this model to predict future traffic.")
        print("Example usage:")
        print("  predictor.load_model()")
        print("  recent_traffic = np.array([...])")  # Last 60 observations
        print("  prediction = predictor.predict(recent_traffic)")
        
    except Exception as e:
        print(f"\nError during training: {str(e)}")
        raise


if __name__ == '__main__':
    main()