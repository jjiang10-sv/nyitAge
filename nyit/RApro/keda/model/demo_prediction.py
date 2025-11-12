"""
Demo script for using the trained traffic prediction model
"""

import numpy as np
#from bentoml.nyitAge.nyit.RApro.keda.model.traffic_predictor import TrafficPredictor
from traffic_predictor import TrafficPredictor
import pandas as pd
from datetime import datetime, timedelta


def demo_prediction():
    """
    Demonstrate loading the model and making predictions
    """
    print("="*60)
    print("TRAFFIC PREDICTION DEMO")
    print("="*60)
    
    # Load the trained model
    predictor = TrafficPredictor()
    
    try:
        predictor.load_model(
            model_path='traffic_model.keras',
            scaler_path='traffic_scaler.pkl'
        )
    except FileNotFoundError:
        print("\nError: Model not found. Please run training first:")
        print("  python traffic_predictor.py")
        return
    
    print(f"\nModel Configuration:")
    print(f"  Time window: {predictor.time_window}")
    print(f"  Lookback steps: {predictor.lookback_steps}")
    print(f"  Prediction horizon: {predictor.prediction_horizon}")
    
    # Simulate recent traffic data
    # In practice, you would get this from your actual monitoring system
    print("\n" + "="*60)
    print("SCENARIO: Predicting traffic based on recent observations")
    print("="*60)
    
    # Example: Generate synthetic recent traffic (in practice, use real data)
    # Simulate a pattern: baseline of 100 requests with some variation
    np.random.seed(42)
    baseline = 100
    recent_traffic = baseline + np.random.normal(0, 20, predictor.lookback_steps)
    recent_traffic = np.maximum(recent_traffic, 0)  # Ensure non-negative
    
    print(f"\nRecent traffic observations (last {predictor.lookback_steps} periods):")
    print(f"  Mean: {recent_traffic.mean():.2f} requests per {predictor.time_window}")
    print(f"  Min: {recent_traffic.min():.2f}")
    print(f"  Max: {recent_traffic.max():.2f}")
    print(f"  Last 5 values: {recent_traffic[-5:]}")
    
    # Make prediction
    print("\n" + "-"*60)
    print("Making prediction...")
    print("-"*60)
    
    prediction = predictor.predict(recent_traffic)
    
    print(f"\nPredicted traffic for next {predictor.prediction_horizon} periods:")
    for i, traffic in enumerate(prediction):
        print(f"  Period {i+1}: {traffic:.2f} requests")
    
    print(f"\nPrediction summary:")
    print(f"  Predicted mean: {prediction.mean():.2f} requests per {predictor.time_window}")
    print(f"  Predicted min: {prediction.min():.2f}")
    print(f"  Predicted max: {prediction.max():.2f}")
    
    # Calculate trend
    trend = prediction.mean() - recent_traffic[-10:].mean()
    if trend > 5:
        print(f"  Trend: INCREASING (↑ {trend:.2f} requests)")
    elif trend < -5:
        print(f"  Trend: DECREASING (↓ {abs(trend):.2f} requests)")
    else:
        print(f"  Trend: STABLE (~ {abs(trend):.2f} requests)")
    
    # Demonstrate real-time scenario
    print("\n" + "="*60)
    print("SCENARIO: Real-time monitoring and alerting")
    print("="*60)
    
    # Define thresholds
    normal_threshold = 150
    high_threshold = 200
    
    print(f"\nThresholds defined:")
    print(f"  Normal: < {normal_threshold} requests")
    print(f"  High: {normal_threshold}-{high_threshold} requests")
    print(f"  Critical: > {high_threshold} requests")
    
    print(f"\nAlert analysis:")
    for i, traffic in enumerate(prediction):
        period_time = datetime.now() + timedelta(minutes=5*(i+1))
        time_str = period_time.strftime("%H:%M")
        
        if traffic > high_threshold:
            print(f"  🚨 CRITICAL: Period {i+1} ({time_str}): {traffic:.0f} requests")
        elif traffic > normal_threshold:
            print(f"  ⚠️  WARNING: Period {i+1} ({time_str}): {traffic:.0f} requests")
        else:
            print(f"  ✓  NORMAL: Period {i+1} ({time_str}): {traffic:.0f} requests")
    
    print("\n" + "="*60)
    print("DEMO COMPLETE")
    print("="*60)
    print("\nUse cases for this model:")
    print("  1. Auto-scaling: Scale servers based on predicted load")
    print("  2. Alerting: Send alerts before traffic spikes")
    print("  3. Resource planning: Plan infrastructure upgrades")
    print("  4. Cost optimization: Reduce resources during low traffic")
    print("  5. Performance monitoring: Detect anomalies in traffic patterns")


def batch_prediction_demo():
    """
    Demonstrate batch prediction for multiple time periods
    """
    print("\n" + "="*60)
    print("BATCH PREDICTION DEMO")
    print("="*60)
    
    predictor = TrafficPredictor()
    
    try:
        predictor.load_model()
    except FileNotFoundError:
        print("\nError: Model not found. Train the model first.")
        return
    
    # Simulate multiple scenarios
    scenarios = {
        'Low Traffic': np.random.normal(50, 10, predictor.lookback_steps),
        'Normal Traffic': np.random.normal(100, 20, predictor.lookback_steps),
        'High Traffic': np.random.normal(200, 30, predictor.lookback_steps),
        'Increasing Trend': np.linspace(50, 150, predictor.lookback_steps),
        'Decreasing Trend': np.linspace(150, 50, predictor.lookback_steps),
    }
    
    results = []
    
    for scenario_name, recent_data in scenarios.items():
        recent_data = np.maximum(recent_data, 0)
        prediction = predictor.predict(recent_data)
        
        results.append({
            'Scenario': scenario_name,
            'Recent Avg': recent_data.mean(),
            'Predicted Avg': prediction.mean(),
            'Change': prediction.mean() - recent_data[-10:].mean()
        })
    
    # Display results
    df_results = pd.DataFrame(results)
    print("\nBatch Prediction Results:")
    print(df_results.to_string(index=False))
    
    print("\n" + "="*60)


if __name__ == '__main__':
    # Run both demos
    demo_prediction()
    print("\n\n")
    batch_prediction_demo()