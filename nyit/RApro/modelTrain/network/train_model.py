import numpy as np
# from sklearn.datasets import make_classification
from sklearn.ensemble import RandomForestClassifier
# from sklearn.model_selection import train_test_split
import joblib
import bentoml
import pandas as pd
from sklearn.metrics import accuracy_score, f1_score
from scipy.stats import zscore


# Load your NSL-KDD training dataset
#data_train = pd.read_csv('KDDTrain+_20Percent_615.csv', header=None)
data_train = pd.read_csv('KDDTrain+_20Percent_615.csv')  # Keep the header

# Load your NSL-KDD testing dataset
data_test = pd.read_csv('KDDTest+_615.csv')


# To do...
# Count the number of data points for Regular (0), DoS (1), R2L (2), U2R (3), Probe (4) in both training and testing datasets
# training dataset
# Count the number of data points for each class in the training dataset
num_regular = data_train[data_train.iloc[:, -1] == 0].shape[0]
num_dos = data_train[data_train.iloc[:, -1] == 1].shape[0]
num_r2l = data_train[data_train.iloc[:, -1] == 2].shape[0]
num_u2r = data_train[data_train.iloc[:, -1] == 3].shape[0]
num_probe = data_train[data_train.iloc[:, -1] == 4].shape[0]
print('\Training Dataset:')
print('Regular data points:', num_regular)
print('DoS data points:', num_dos)
print('R2L data points:', num_r2l)
print('U2R data points:', num_u2r)
print('Probe data points:', num_probe)

# Count the number of data points for each class in the testing dataset
num_regular_test = data_test[data_test.iloc[:, -1] == 0].shape[0]
num_dos_test = data_test[data_test.iloc[:, -1] == 1].shape[0]
num_r2l_test = data_test[data_test.iloc[:, -1] == 2].shape[0]
num_u2r_test = data_test[data_test.iloc[:, -1] == 3].shape[0]
num_probe_test = data_test[data_test.iloc[:, -1] == 4].shape[0]

print('\nTesting Dataset:')
print('Regular data points:', num_regular_test)
print('DoS data points:', num_dos_test)
print('R2L data points:', num_r2l_test)
print('U2R data points:', num_u2r_test)
print('Probe data points:', num_probe_test)

categorical_features = ["protocol_type", "service", "flag"]
train_df = data_train.copy()
test_df = data_test.copy()

for feature in categorical_features:
    train_dummies = pd.get_dummies(train_df[feature], prefix=feature)

    #import pdb; pdb.set_trace()
    # Get the columns to ensure alignment with test data
    feature_columns = train_dummies.columns

    # Generate dummy variables for the test data
    test_dummies = pd.get_dummies(test_df[feature], prefix=feature)
    # Reindex test dummy columns to match training columns, filling missing with 0
    test_dummies = test_dummies.reindex(columns=feature_columns, fill_value=0)

    # Add dummies to dataframe using a different method
    for col in train_dummies.columns:
        train_df[col] = train_dummies[col]
        test_df[col] = test_dummies[col]
    train_df = train_df.drop(columns=[feature])
    test_df = test_df.drop(columns=[feature])



import numpy as np #import numpy
from scipy.stats import zscore
# Get training and test data and labels for model training
# Convert pandas DataFrames to NumPy arrays
features_train = train_df.drop(columns=["labels"]).values  # Exclude last column (target)
labels_train = train_df['labels'].values  # Target variable in the last column
features_test = test_df.drop(columns=["labels"]).values  # Exclude last column (target)
labels_test = test_df['labels'].values  # Target variable in the last column

features_train = np.nan_to_num(features_train).astype(np.float64)  # Replace NaNs, convert to float64
features_test = np.nan_to_num(features_test).astype(np.float64)  # Replace NaNs, convert to float64

# Normalize the training and test datasets using zscore
features_train = zscore(features_train, axis=0, ddof=1)
features_test = zscore(features_test, axis=0, ddof=1)

# Import the Python libraries
import time
# Create and train your model
time_start = time.time()  # training time - start
#model = DecisionTreeClassifier()  # Initialize the DecisionTreeClassifier
# Initialize RandomForestClassifier with parameters
model = RandomForestClassifier(n_estimators=200, 
                               max_depth=10, 
                               random_state=42,
                               class_weight='balanced'
                               )

# Generate the model using training data and labels
model.fit(features_train, labels_train)

time_end = time.time()  # training time - end
training_time = time_end - time_start
print('Training completed')
print('Training time:', training_time)

# Import the Python libraries
from sklearn.metrics import accuracy_score
from sklearn.metrics import f1_score

# Testing, for sklearn libriary if applicable
predicted_labels = model.predict(features_test)
accuracy = accuracy_score(labels_test, predicted_labels)
fscore = f1_score(labels_test, predicted_labels, average="weighted")
# Show the results: accuracy and training time
print('Accuracy:', accuracy)
print('F1-Score:', fscore)

# ... (Previous code for data loading, preprocessing, and model training) ...

# 1. Feature Selection:
# Get feature importances from the trained model
feature_importances = model.feature_importances_

# Get the indices of the top 5 features
top_10_feature_indices = np.argsort(feature_importances)[-10:]

# Get the names of the top 5 features (if you have feature names)
# Assuming 'train_df' has the feature names as columns
top_10_feature_names = train_df.columns[top_10_feature_indices]

print("Top 10 important features:", top_10_feature_names)

# 2. Re-run with Selected Features:
# Create new training and testing datasets with only the top 5 features
features_train_selected = features_train[:, top_10_feature_indices]
features_test_selected = features_test[:, top_10_feature_indices]

# Re-train the model with selected features
#model_selected = DecisionTreeClassifier()
model_selected = RandomForestClassifier(n_estimators=200, max_depth=10, random_state=42,class_weight="balanced")
model_selected.fit(features_train_selected, labels_train)

# 3. Re-calculate Metrics and Compare:
# Make predictions on the test set with the selected features
predicted_labels_selected = model_selected.predict(features_test_selected)

# Calculate Accuracy and F1-Score for the selected features model
accuracy_selected = accuracy_score(labels_test, predicted_labels_selected)
f1_selected = f1_score(labels_test, predicted_labels_selected, average='weighted')

# Print and compare the results
print("\nOriginal Model:")
print("Accuracy:", accuracy)
print("F1-Score:", fscore)

print("\nSelected Features Model:")
print("Accuracy:", accuracy_selected)
print("F1-Score:", f1_selected)



# Compare the metrics and analyze the impact of feature selection
# Save model to BentoML Model Store
with bentoml.models.create(name='intrusion_detector') as model_ref:
    joblib.dump(model_selected, model_ref.path_of("intrusion_detector_model.pkl"))
    print(f"Model saved: {model_ref}")
