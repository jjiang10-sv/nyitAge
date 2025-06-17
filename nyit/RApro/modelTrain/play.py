from sklearn.preprocessing import StandardScaler, MinMaxScaler
import numpy as np

# Example data
data = np.array([[100, 2], [200, 4], [300, 6], [400, 8]])

scaler = StandardScaler()
standardized = scaler.fit_transform(data)

scaler = MinMaxScaler()
minmaxed = scaler.fit_transform(data)

print("Original data:")
print(data)
print("\nStandardized data:")
print(standardized)
print(f"\nMean: {standardized.mean(axis=0)}")  # ~[0, 0]
print(f"Std: {standardized.std(axis=0)}")     # ~[1, 1]
print("\nMinMaxScaled data:")
print(minmaxed)
print(f"\nMin: {minmaxed.min(axis=0)}")
print(f"Max: {minmaxed.max(axis=0)}")