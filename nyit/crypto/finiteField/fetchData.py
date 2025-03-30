# Here's the Python equivalent using the `requests` library with all headers included:

# ```python
import requests
import pandas as pd
import re
from pprint import pprint
from datetime import datetime

# fetch("https://secure.workforceready.com.au/ta/reports/export/12914898908/HTML/DetailedHoursOverview-SYD_1742094442663.html?ActiveSessionId=5623878805&DoHead=1&UseUnicode=0&ReqNum=$8221779032", {
#   "headers": {
#     "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
#     "accept-language": "zh-CN,zh;q=0.9",
#     "cache-control": "max-age=0",
#     "priority": "u=0, i",
#     "sec-ch-ua": "\"Chromium\";v=\"134\", \"Not:A-Brand\";v=\"24\", \"Google Chrome\";v=\"134\"",
#     "sec-ch-ua-mobile": "?0",
#     "sec-ch-ua-platform": "\"macOS\"",
#     "sec-fetch-dest": "document",
#     "sec-fetch-mode": "navigate",
#     "sec-fetch-site": "same-origin",
#     "sec-fetch-user": "?1",
#     "upgrade-insecure-requests": "1",
#     "cookie": "c0cef3b930ae64e722d0db285d41a3598e20bc5ca4c1059156b20ff2d5998e46=11e5e6eeaea492d79ff4559bdaf058f626c7259ab12acefe60fe45bbef3aa3d1065ff082f472f0d4be1f52849f9f901a511cd499b63349c4; JSESSIONID=C3D58BD6975CDC09D3FAE844B176CD45; XSRF-TOKEN=FOQOZTOOEZ; LastLoginTime=\"1742094194629| 2025-03-15 23:03:14\"; lbSession=47b67c9f554f22154c903461d1a3d606; _dd_s=",
#     "Referer": "https://secure.workforceready.com.au/",
#     "Referrer-Policy": "strict-origin"
#   },
#   "body": null,
#   "method": "GET"
# }); ;

# fetch("https://secure.workforceready.com.au/ta/reports/export/12914811017/HTML/DetailedHoursOverview-SYD_1741836107025.html?ActiveSessionId=5621944127&DoHead=1&UseUnicode=0&ReqNum=$6555361241", {
# "headers": {
# "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
# "accept-language": "zh-CN,zh;q=0.9",
# "cache-control": "max-age=0",
# "priority": "u=0, i",
# "sec-ch-ua": "\"Not(A:Brand\";v=\"99\", \"Google Chrome\";v=\"133\", \"Chromium\";v=\"133\"",
# "sec-ch-ua-mobile": "?0",
# "sec-ch-ua-platform": "\"macOS\"",
# "sec-fetch-dest": "document",
# "sec-fetch-mode": "navigate",
# "sec-fetch-site": "same-origin",
# "sec-fetch-user": "?1",
# "upgrade-insecure-requests": "1",
# "cookie": "JSESSIONID=44A6C071615CC2A2BBEA5F2FDD202969; XSRF-TOKEN=TQWJAWFTZI; LastLoginTime=\"1741835929981| 2025-03-12 23:18:49\"; lbSession=da78e1f8f77b0f53556cfc1c8fd446ec; _dd_s=logs=1&id=7d67dc3e-3393-4014-8c8f-6c681ef401bd&created=1741838816084&expire=1741840627319&rum=0",
# "Referer": "https://secure.workforceready.com.au/",
# "Referrer-Policy": "strict-origin"


url = "https://secure.workforceready.com.au/ta/reports/export/12914898908/HTML/DetailedHoursOverview-SYD_1742094442663.html?ActiveSessionId=5623878805&DoHead=1&UseUnicode=0&ReqNum=$8221779032"

headers = {
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
}

def req_func():
    response = requests.get(url, headers=headers)
    if response.status_code == 200:
        # Process the HTML content
        html_content = response.text
        
        # Example: Parse tables with pandas
        
        tables = pd.read_html(html_content)
        
        # Save first table to CSV
        if len(tables) > 0:
            df = tables[0]
            df.to_csv('report_data.csv', index=False)
            print("Data saved to report_data.csv")
        else:
            print("No tables found in the HTML content")
    else:
        print(f"Request failed with status code: {response.status_code}")
        print(response.text)

def convert_time_to_int_min(time_str):
    hours, minutes = map(int, time_str.split(':'))
    return hours*60 + minutes

def calculate_time_differences(times):
    time_a = []
    time_p = []
    # Separate times into 'a' and 'p'
    for time in times:
        if time.endswith('a'):
            time_a.append(time[:-1])  # Remove 'a'
        elif time.endswith('p'):
            time_p.append(time[:-1])  # Remove 'a'
    # Calculate differences
    a_diff = convert_time_to_int_min(time_a[1])-convert_time_to_int_min(time_a[0])
    p_diff = convert_time_to_int_min(time_p[1])-convert_time_to_int_min(time_p[0])
    return a_diff, p_diff


def read_data():
    # Read the CSV saved from the HTML table
    # Load your data
    df = pd.read_csv('report_data.csv')

    week_days = ["Sun","Mon","Tue","Wed","Thu","Fri"]

    ot, lateA, earlyP, noonBreak = "overtime","late_a","early_p", "noon_break"
    cal_t, raw_t, total_w, schedule_d = "_cal_t","_raw_t","_total_w", "_schedule"

    for day in  week_days:
        # df[day+ot] = None
        # df[day+lateA] = None
        # df[day+earlyP] = None
        # df[day+noonBreak] = None
        df[day+cal_t] = None
        df[day+raw_t] = None
        df[day+schedule_d] = None
    df[total_w] = None
    df_len = len(df)

    for i in range (2,df_len):
        text = df.iloc[i, 4]
        #text = 'Sun 09-03Mon 10-03Sch:06:00a-02:00p\xa004:59a-\xa002:00p(Warehouse fo)Calc Total:8:36Raw Total:9:01Tue 11-03Sch:06:00a-02:00p\xa005:03a-\xa002:03p(Warehouse fo)Calc Total:8:06Raw Total:9:00Wed 12-03Sch:06:00a-02:00p\xa006:01a-\xa002:01p(Warehouse fo)Calc Total:7:36Raw Total:8:00Thu 13-03Sch:06:00a-02:00p\xa005:02a-\xa002:02p(Warehouse fo)Calc Total:8:06Raw Total:9:00TotalCalc TimeTotalA & J Austra/NSW/Warehouse/Warehouse fo32:2432:24Total32:2432:24'

        # Clean up non-breaking spaces and create consistent separators
        if pd.isna(text):
            continue
        clean_text = text.replace('\xa0', ' ').replace('Sch:', ' Sch:').replace('Calc', ' Calc')
        day_entries = re.findall(r'(Sun|Mon|Tue|Wed|Thu|Fri|Sat)\s(\d{2}-\d{2})(.*?)(?=(?:Sun|Mon|Tue|Wed|Thu|Fri|Sat)\s\d{2}-\d{2}|Total Calc TimeTotal)', clean_text, re.DOTALL)

        # Split totals section
        totals_section = re.search(r'Total Calc TimeTotal(.*)$', clean_text).group(1) if 'Total Calc TimeTotal' in clean_text else None
        # Process daily entries
        schedule_data = []
        #result_df = 
        for day in day_entries:
            day_name, date, content = day[0], day[1], day[2].strip()  
            # Extract components
            entry = {
                'Day': day_name,
                'Date': date,
                #'Schedule': re.search(r'Sch:(.*?) Calc', content).group(1).strip() if ' Calc' in content else content,
                'Schedule': content.split("(")[0] if '(' in content else content,
                'Location': re.search(r'\((.*?)\)', content).group(1) if '(' in content else None,
                'Calc Total': re.search(r'Calc Total:([\d:]+)', content).group(1) if 'Calc Total' in content else None,
                'Raw Total': re.search(r'Raw Total:([\d:]+)', content).group(1) if 'Raw Total' in content else None
            }
            col_day_name = entry["Day"]
            schedule = entry["Schedule"]
            # time_entries = []
            # time_differences = []
            # if schedule is not None:
            #     time_entries = re.findall(r'(\d{2}:\d{2}[ap])', schedule)
            # if len(time_entries) == 4 :
            #     diff_a, diff_p = calculate_time_differences(time_entries)
            #     if diff_a > 0:
            #         df[i, col_day_name+lateA] = diff_a
            #     if diff_p < 0 :
            #         df[i, col_day_name+earlyP] = diff_a
            #     df[i, col_day_name+ot] = diff_a - diff_p

            df.at[i, col_day_name+raw_t] = entry["Raw Total"]
            df.at[i, col_day_name+cal_t] = entry["Calc Total"]
            
            df.at[i, col_day_name+schedule_d] = schedule
            
            #schedule_data.append(entry)

        # Process totals
        totals = {
            'Total Calc Time':  re.findall(r'\d{2}:\d{2}', totals_section) if totals_section else [None],
            'Organization': re.search(r'(.*?)\d', totals_section).group(1).strip('/') if totals_section else None,
            #'Values': re.findall(r'\d{2}:\d{2}', totals_section)
        }
        df.at[i, total_w] = totals["Total Calc Time"][0]

        # print("Daily Schedule Entries:")
        # pprint(schedule_data)

        # print("\nTotals Section:")
        # pprint(totals)
        # print("\nRow 0 values:")
    df.drop([0,1])
    #df = df.drop(['Column1', 'Column2'], axis=1)  # Remove specified columns
    df = df.drop(df.columns[[4]], axis=1)  # Removes 'Column1' and 'Column2'

    df.to_csv('report_data_process.csv', index=False)

req_func()
read_data()
#if __name__ == "main":
# ```

# Key points:
# 1. The `headers` dictionary includes all the required security headers and cookies
# 2. Uses `requests.get()` with the exact same parameters as the JavaScript fetch
# 3. Includes error handling and status code checking
# 4. Uses pandas to parse HTML tables (common format for reports)
# 5. Saves the first found table to CSV

# Important notes:
# - The session cookies (JSESSIONID, XSRF-TOKEN, etc.) will eventually expire
# - The ActiveSessionId in the URL parameter is likely temporary
# - You may need to handle authentication first if the session expires
# - The cookie values should be treated as sensitive credentials
# - Add timeouts and retry logic for production use

# To inspect the raw HTML content first, you can add:
# ```python
# with open('raw_page.html', 'w') as f:
#     f.write(html_content)
# ```