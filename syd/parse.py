import pandas as pd

# Load the Excel file
file_path = "Detailed Hours Overview.xlsx"  # Update with your file path
xls = pd.ExcelFile(file_path)

# Load the first sheet, skipping metadata rows
df = pd.read_excel(xls, sheet_name=xls.sheet_names[0], skiprows=4)

# Set the correct header row
df.columns = df.iloc[0]  # Assign first row as column names
df = df[1:].reset_index(drop=True)  # Remove the old header row

# Drop completely empty columns
df = df.dropna(axis=1, how='all')

# Display the first few rows
print(df.head())
