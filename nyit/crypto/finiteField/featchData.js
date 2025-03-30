fetch("https://secure.workforceready.com.au/ta/reports/export/12914898908/HTML/DetailedHoursOverview-SYD_1742094442663.html?ActiveSessionId=5623878805&DoHead=1&UseUnicode=0&ReqNum=$8221779032", {
    "headers": {
      "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
      "accept-language": "zh-CN,zh;q=0.9",
      "cache-control": "max-age=0",
      "priority": "u=0, i",
      "sec-ch-ua": "\"Chromium\";v=\"134\", \"Not:A-Brand\";v=\"24\", \"Google Chrome\";v=\"134\"",
      "sec-ch-ua-mobile": "?0",
      "sec-ch-ua-platform": "\"macOS\"",
      "sec-fetch-dest": "document",
      "sec-fetch-mode": "navigate",
      "sec-fetch-site": "same-origin",
      "sec-fetch-user": "?1",
      "upgrade-insecure-requests": "1",
      "cookie": "c0cef3b930ae64e722d0db285d41a3598e20bc5ca4c1059156b20ff2d5998e46=11e5e6eeaea492d79ff4559bdaf058f626c7259ab12acefe60fe45bbef3aa3d1065ff082f472f0d4be1f52849f9f901a511cd499b63349c4; JSESSIONID=C3D58BD6975CDC09D3FAE844B176CD45; XSRF-TOKEN=FOQOZTOOEZ; LastLoginTime=\"1742094194629| 2025-03-15 23:03:14\"; lbSession=47b67c9f554f22154c903461d1a3d606; _dd_s=",
      "Referer": "https://secure.workforceready.com.au/",
      "Referrer-Policy": "strict-origin"
    },
    "body": null,
    "method": "GET"
  }); ;
//     To create a pandas DataFrame and set row/column data, follow these steps:

// ### 1. Import pandas
// ```python
// import pandas as pd
// ```

// ### 2. Create a DataFrame
// #### Method 1: From a dictionary (keys = columns)
// ```python
// data = {
//     'Name': ['Alice', 'Bob', 'Charlie'],
//     'Age': [25, 30, 35],
//     'City': ['New York', 'London', 'Tokyo']
// }

// df = pd.DataFrame(data)
// print(df)
// ```

// #### Method 2: From a list of lists (specify columns)
// ```python
// data = [
//     ['Alice', 25, 'New York'],
//     ['Bob', 30, 'London'],
//     ['Charlie', 35, 'Tokyo']
// ]

// df = pd.DataFrame(data, columns=['Name', 'Age', 'City'])
// print(df)
// ```

// #### Method 3: From a list of dictionaries
// ```python
// data = [
//     {'Name': 'Alice', 'Age': 25, 'City': 'New York'},
//     {'Name': 'Bob', 'Age': 30, 'City': 'London'},
//     {'Name': 'Charlie', 'Age': 35, 'City': 'Tokyo'}
// ]

// df = pd.DataFrame(data)
// print(df)
// ```

// ### 3. Set Column Data
// #### Add a new column
// ```python
// df['Salary'] = [70000, 80000, 90000]  # Length must match row count
// ```

// #### Modify existing column
// ```python
// df['Age'] = [26, 31, 36]  # Update all values
// ```

// ### 4. Set Row Data
// #### Add a new row (modern method using concat)
// ```python
// new_row = pd.DataFrame({'Name': ['Dave'], 'Age': [28], 'City': ['Paris']})
// df = pd.concat([df, new_row], ignore_index=True)
// ```

// #### Modify existing row
// ```python
// # Using index
// df.loc[0] = ['Alicia', 26, 'Los Angeles', 75000]  # Modify first row

// # Using conditional update
// df.loc[df['Name'] == 'Bob', 'City'] = 'Manchester'
// ```

// ### 5. Set Custom Index
// ```python
// df = pd.DataFrame(data, index=['a', 'b', 'c'])  # During creation
// df.index = ['Row1', 'Row2', 'Row3']  # After creation
// ```

// ### Example Output
// ```
//       Name  Age         City  Salary
// 0    Alice   25     New York   70000
// 1      Bob   30      London    80000
// 2  Charlie   35       Tokyo   90000
// 3     Dave   28       Paris     NaN
// ```

// ### Key Notes:
// - Use `pd.DataFrame()` to initialize a DataFrame
// - Columns can be added/modified using dictionary-style assignment
// - Use `pd.concat()` instead of deprecated `append()`
// - `loc[]` is used for label-based indexing
// - `ignore_index=True` resets the index when adding rows
// - All new data must align with existing DataFrame structure

// This covers the most common ways to create and manipulate DataFrames in pandas. Let me know if you need more specific operations!


// # 1. **`re.findall()`**: This function searches the string `clean_text` for all occurrences that match the given regular expression pattern. It returns a list of tuples, where each tuple contains the matched groups.

// # 2. **Regular Expression Pattern**: 
// #    - **`(Sun|Mon|Tue|Wed|Thu|Fri|Sat)`**: This part matches any day of the week. It uses a non-capturing group to match one of the specified strings (e.g., "Sun", "Mon", etc.).
// #    - **`\s`**: Matches a whitespace character (like a space).
// #    - **`(\d{2}-\d{2})`**: Matches a date in the format "dd-dd", where `\d{2}` matches exactly two digits.
// #    - **`(.*?)`**: Matches any character (except a newline) zero or more times, as few times as possible (non-greedy match). This captures the content following the date.
// #    - **`(?=(?:Sun|Mon|Tue|Wed|Thu|Fri|Sat)\s\d{2}-\d{2}|Raw Total)`**: This is a lookahead assertion. It ensures that the match is followed by either another day and date (e.g., "Mon 10-03") or the string "Raw Total". The `(?: ... )` is a non-capturing group used within the lookahead.

// # 3. **`re.DOTALL`**: This flag allows the `.` in the regular expression to match newline characters as well. This is useful when you want to match across multiple lines.

// # ### Purpose:

// # The purpose of this regular expression is to extract entries from `clean_text` that start with a day of the week and a date, followed by any content, and end just before the next day entry or the "Raw Total" section. This is likely used to parse a schedule or log that is organized by days, capturing the details for each day separately.


// # Get rows 0-1 and columns 0-2
//     subset = df.iloc[0:2, 0:2]
//     print("\nSubset:")
//     print(subset)

//     # First 5 rows
//     print("First 5 rows:")
//     print(df.head())

//     # Last 5 rows
//     print("\nLast 5 rows:")
//     print(df.tail())

//     # DataFrame structure
//     print("\nDataFrame info:")
//     print(df.info())

//     # Basic statistics
//     print("\nDescriptive statistics:")
//     print(df.describe(include='all'))  # 'all' includes categorical data
//     # Total missing values per column
//     missing_values = df.isnull().sum()
//     print("Missing values per column:")
//     print(missing_values)

//     # Percentage of missing values
//     print("\nMissing value percentages:")
//     print((df.isnull().sum() / len(df) * 100).round(2))

//     # Visualize missing values (requires matplotlib)
//     # import matplotlib.pyplot as plt
//     # plt.figure(figsize=(10,6))
//     # plt.title('Missing Values Heatmap')
//     # sns.heatmap(df.isnull(), cbar=False)
//     # plt.show()
//     # # For categorical columns
//     # categorical_cols = df.select_dtypes(include=['object']).columns
//     # for col in categorical_cols:
//     #     print(f"\nUnique values in {col}:")
//     #     print(df[col].unique()[:10])  # First 10 unique values
//     #     print(f"Total unique: {df[col].nunique()}")
//     # print(report_df)

//     # # If you need to handle special characters:
//     # report_df = pd.read_csv('report_data.csv', encoding='utf-8-sig')

//     # # If numeric columns have formatting ($, commas):
//     # report_df = pd.read_csv('report_data.csv', thousands=',', decimal='.')
