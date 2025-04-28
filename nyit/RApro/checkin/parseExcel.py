import pandas as pd
import json
import os

def parse_excel_to_dict(file_path, col1, col2):
    # Read the Excel file
    df = pd.read_excel(file_path)

    # Check if the specified columns exist
    if col1 not in df.columns or col2 not in df.columns:
        raise ValueError(f"Columns '{col1}' or '{col2}' not found in the Excel sheet.")

    # Create a dictionary mapping from the two columns with keys as integers, excluding NaN keys
    result_dict = {int(key): value for key, value in zip(df[col1], df[col2]) if pd.notna(key)}

    return result_dict


# Example usage
if __name__ == "__main__":
    file_path = 'SYDPayroll.xlsx'  # Path to your Excel file
    column1 = 'EM NO.'      # Name of the first column
    column2 = 'col1'      # Name of the second column

    try:
        mapped_dict = parse_excel_to_dict(file_path, column1, column2)
         # Check if the file exists before writing
        if os.path.exists('mapped_dict.py'):
            os.remove('mapped_dict.py')  # Remove the file if it exists
        
        # Save the dictionary to a Python file
        with open('mapped_dict.py', 'w') as f:
            f.write(f"mapped_dict = {json.dumps(mapped_dict, indent=4,default=int)}\n")
        
        
        #parse_excel_to_dict_list(file_path,column1, column2)
    except Exception as e:
        print(f"An error occurred: {e}")