import pandas as pd
from bs4 import BeautifulSoup
import ast
import csv

official_worker = "F/T"
mapped_dict = {
    1289: "CAS",
    1145: "F/T",
    9422: "F/T",
    10098: "CAS",
    9943: "F/T",
    10184: "F/T",
    1319: "CAS",
    9140: "F/T",
    9131: "F/T",
    9427: "F/T",
    1310: "CAS",
    1271: "CAS",
    1203: "F/T",
    1147: "CAS",
    1266: "CAS",
    1293: "CAS",
    8992: "F/T",
    1221: "CAS",
    1200: "CAS",
    10102: "F/T",
    1178: "F/T",
    1162: "CAS",
    1202: "CAS",
    1230: "CAS",
    1287: "CAS",
    1295: "CAS",
    10186: "F/T",
    9092: "F/T",
    1170: "CAS",
    1317: "F/T",
    9491: "F/T",
    9888: "F/T",
    9141: "F/T",
    1321: "CAS",
    9606: "F/T",
    1154: "F/T",
    1232: "CAS",
    1284: "CAS",
    9176: "F/T",
    10188: "F/T",
    10021: "CAS",
    1196: "F/T",
    9188: "F/T",
    1300: "CAS",
    1302: "CAS",
    10230: "F/T",
    10088: "CAS",
    1312: "CAS",
    9355: "F/T",
    1256: "CAS",
    1315: "CAS",
    1137: "CAS",
    1299: "F/T",
    9572: "F/T",
    1263: "CAS",
    1199: "F/T",
    10101: "F/T",
    1201: "F/T",
    1297: "F/T",
    9944: "F/T",
    1255: "CAS",
    9468: "F/T",
    1198: "F/T",
    1296: "F/T",
    1324: "CAS",
    1260: "F/T",
    10024: "CAS",
    1313: "F/T",
    9913: "F/T",
    1308: "CAS",
    9937: "F/T",
    9887: "F/T",
    9094: "F/T",
    1223: "CAS",
    1291: "CAS",
    1309: "CAS",
    1267: "CAS",
    1272: "CAS",
    9499: "F/T",
    1278: "CAS",
    1183: "F/T",
    9519: "F/T",
    1224: "F/T",
    9500: "F/T",
    1298: "CAS",
    1273: "CAS",
    1290: "CAS",
    1123: "F/T",
    1306: "CAS",
    1276: "CAS",
    1307: "CAS",
    9108: "F/T",
    9096: "F/T",
    1305: "CAS",
    1286: "CAS",
    9158: "F/T",
    9093: "F/T",
    9948: "F/T"
}

# Get the keys
mapped_dict_keys = mapped_dict.keys()
# Convert to a list if needed
keys_list = list(mapped_dict_keys)


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


def convert_time_to_int_min(time_str):
    hours, minutes = map(int, time_str.split(':'))
    return hours*60 + minutes

def clean_text(date_text):
    tmp = date_text.split("\n")
    tmp1 = None
    if len(tmp) > 1 :
        # seems like only in scheduled_start and end
        tmp1 = tmp[1].strip().split(" ")
    elif len(tmp) == 1:
        tmp1 = tmp[0].strip().split(" ")
    if len(tmp1) > 1:
        return tmp1[1]
    else:
        return tmp1[0]

def extra_data():

    try:
        with open('detailsHours1.htm', 'r', encoding='utf-8') as file:
            html_content = file.read()

        soup = BeautifulSoup(html_content, 'html.parser')
        main_table = soup.find('table', {'class': 'reportTable'})
        rows = main_table.find_all('tr', class_=lambda x: x and x.startswith('resultRow'))

        data = []
        for row in rows:
            row_data = []
            columns = row.find_all('td')
            if len(columns) < 5:
                continue
            
            emp_id = columns[0].get_text(strip=True)
            if emp_id in ["10102"]:
                print(emp_id)
            name = columns[1].get_text(strip=True)
            department = columns[2].get_text(strip=True)
            status = columns[3].get_text(strip=True)
            weekly_overview = columns[4]
            
            # Extract daily details
            days = weekly_overview.find_all('td', class_=lambda x: x and ('wddTL' in x or 'wddTLR' in x))
            employ_week_schedule = {
                'Employee ID': emp_id,
                'Name': name,
                'Department': department,
                'Status': status,
            }
            for day in days:
                schedules_actives = []
                date_header = day.find('div', class_='DateHeader')
                date = date_header.get_text(strip=True) if date_header else None
                if date is not None:

                
                    # Extract schedule times
                    schedule_table = day.find('table', class_='DailyDetails')
                    schedule_start = schedule_end = None
                    if schedule_table:
                        sch_tds = schedule_table.find_all('td')
                        if len(sch_tds) >= 4:
                            schedule_start = clean_text(sch_tds[1].get_text(strip=True))
                            schedule_end = clean_text(sch_tds[3].get_text(strip=True))
                    
                    # Extract actual clock-in/out times
                    actual_tables = day.find_all('table', class_='DailyDetails')
                    actual_start = actual_end = None
                    actual_start_1 = actual_end_1 = None

                    carryover_start = carryover_end = None
                    carryover_start_1 = carryover_end_1 = None
                    if len(actual_tables) > 1:
                        actual_tds = actual_tables[1].find_all('td')
                        # actual_tds[9].get_text(strip=True)    process work
                        if len(actual_tds) >= 9:  # Adjusted index based on HTML structure
                            actual_start = clean_text(actual_tds[4].get_text(strip=True))
                            # tmp = actual_start.split("\n")
                            # if len(tmp) > 1 :
                            #     actual_start = tmp[1].strip()
                            actual_end = clean_text(actual_tds[7].get_text(strip=True))
                        if len(actual_tds) >= 20:  # Adjusted index based on HTML structure
                            actual_start_1 = clean_text(actual_tds[14].get_text(strip=True))
                            actual_end_1 = clean_text(actual_tds[17].get_text(strip=True))
                        
                        if len(actual_tds) >= 40:
                            carryover_start = clean_text(actual_tds[24].get_text(strip=True))
                            carryover_end = clean_text(actual_tds[27].get_text(strip=True))
                            carryover_start_1 = clean_text(actual_tds[34].get_text(strip=True))
                            carryover_end_1 = clean_text(actual_tds[37].get_text(strip=True))

                    elif len(actual_tables) == 1:
                        actual_tds = actual_tables[0].find_all('td')
                        # actual_tds[9].get_text(strip=True)    process work
                        if len(actual_tds) >= 9:  # Adjusted index based on HTML structure
                            actual_start = clean_text(actual_tds[4].get_text(strip=True))
                            # tmp = actual_start.split("\n")
                            # if len(tmp) > 1 :
                            #     actual_start = tmp[1].strip()
                            actual_end = clean_text(actual_tds[7].get_text(strip=True))
                        if len(actual_tds) >= 20:  # Adjusted index based on HTML structure
                            actual_start_1 = clean_text(actual_tds[14].get_text(strip=True))
                            actual_end_1 = clean_text(actual_tds[17].get_text(strip=True))
                        employ_week_schedule[date] = [schedule_start,schedule_start,actual_start,actual_end,actual_start_1,actual_end_1,calc_daily,raw_daily]
                    total_tables = day.find_all('table', class_='TableTotal')
                    calc_daily = total_tables[0].find('td', class_='Data').get_text(strip=True) if len(total_tables) > 0 else None
                    raw_daily = total_tables[1].find('td', class_='Data').get_text(strip=True) if len(total_tables) > 1 else None
                    employ_week_schedule[date] = [schedule_start,schedule_end,actual_start,actual_end,actual_start_1,actual_end_1,calc_daily,raw_daily]
                    if carryover_start is not None:
                        employ_week_schedule[date].append(carryover_start)
                        employ_week_schedule[date].append(carryover_end)
                        employ_week_schedule[date].append(carryover_start_1)
                        employ_week_schedule[date].append(carryover_end_1)
            # Extract weekly totals
            table_keymap2 = weekly_overview.find('table', {'class': 'TableKeyMap2'})
            if table_keymap2:
                data_row = table_keymap2.find('tr', {'class': 'Data1'})
                if data_row:
                    calc_time = data_row.find('td', {'class': 'Data1'}).get_text(strip=True)
                    total_hours = data_row.find('td', {'class': 'RowTotal'}).get_text(strip=True)
                    employ_week_schedule["weeklyCalcTime"] = calc_time
                    employ_week_schedule["weeklyTotalHours"] = total_hours
            data.append(employ_week_schedule)

        # Create DataFrame and handle missing columns
        df = pd.DataFrame(data)
        df.to_csv('report_data.csv', index=False)
        #print(df)
    except Exception as e:
        print(e)
def remove_a_p_in_time(time_str):
    return time_str[:-1]

def end_with_a(time_str):
    return time_str not in (None, "") and time_str[-1] == "a"

def end_with_p(time_str):
    return time_str[:-1] == "p"


def convert_time_to_int_min(time_str):
    time_str = remove_a_p_in_time(time_str)
    hours, minutes = map(int, time_str.split(':'))
    return hours*60 + minutes

def time_diff(time_a,time_b, bigger = True):
    # not schduled
    if time_a == "" and time_b == "":
        return 0
    # no leave or come record. not count into time
    if time_b == "?" or time_a == "?":
        return 0
    diff = convert_time_to_int_min(time_b) -convert_time_to_int_min(time_a)
    if bigger and diff < 0 :
        return diff + 12 *60
    return diff

def calculate_time(times_list):
    no_lunch_record = False
    scheduled_time = active_time  = active_schedule_diff =0
    carry_over_time  = None
    lunch_break = come_late_mins = left_early_mins = 0
    #come_early = leave_late = 0
    scheduled_start = times_list[0]
    scheduled_end = times_list[1]
    actual_start = times_list[2]
    time_list_len = len(times_list)
     
    if  scheduled_start is not None and actual_start is not None:
        
        scheduled_time = time_diff(scheduled_start,scheduled_end)
        #if personal leave, then an officer
        if actual_start == "" and times_list[3] == "7:36":
            active_schedule_diff = None
            return active_schedule_diff,come_late_mins, lunch_break, left_early_mins ,carry_over_time
            
        else:
            if time_list_len >= 8 :
                if times_list[4] is None:
                    no_lunch_record = True
                    active_time = time_diff(actual_start,times_list[3])
                else:
                    active_time_1 = time_diff(actual_start,times_list[3])
                    active_time_2 = time_diff(times_list[4],times_list[5])
                    active_time = active_time_1+active_time_2 + 30
                    # not count the lunch break
                    record_time_shift = time_diff(actual_start,times_list[5])
                    if active_time > record_time_shift:
                        active_time = record_time_shift

            if time_list_len > 8:
                
                last_record = times_list[11]
                # no lunch record
                if last_record == '?':
                    carry_over_time = time_diff(times_list[8],times_list[10])
                    carry_over_time = round(carry_over_time/60,1)
                    if carry_over_time > 7.6:
                        carry_over_time = 7.6
                            
                    # no lunch record, deduct time
                    carry_over_time -= 0.25
                            
                else:
                    active_time_3 = time_diff(times_list[8],times_list[9])
                    active_time_4 = time_diff(times_list[10],last_record)
                    carry_over_time = active_time_3 + active_time_4
                    if carry_over_time > 7.6:
                        carry_over_time = 7.6

                    
    # scheduled_time = round(scheduled_time/60,2)
    # active_time = round(active_time/60,2)
    # ot = round(active_time-scheduled_time,2)
        active_schedule_diff = active_time-scheduled_time
        # no scheduled; no personal leave
        if scheduled_start != "" and actual_start != "":
            come_late_mins = time_diff(scheduled_start,actual_start,bigger=False)
        # personal leave
        #print("debug--------", actual_start)
        if no_lunch_record: 
            #to_deduct_hr = deduct_time(to_deduct_hr,miss_mins)
            # no scheduled;
            if scheduled_start != "":
                left_early_mins = time_diff(times_list[3],scheduled_end,bigger=False)
        else:
            
            lunch_break = time_diff(times_list[3],times_list[4])
            # not adding the 30 mins lunch break since active time not count lunch break
            # active_schedule_diff += 30
            # if lunch_break > 30:
            #     to_deduct_hr += 0.25
            #come_late_mins = time_diff(scheduled_start,actual_start,bigger=False)
            #to_deduct_hr = deduct_time(to_deduct_hr,come_late_mins)
            if scheduled_end:
                left_early_mins = time_diff(times_list[5],scheduled_end,bigger=False)
            #to_deduct_hr = deduct_time(to_deduct_hr,left_early_mins)
    
        
    return active_schedule_diff, come_late_mins, lunch_break, left_early_mins ,carry_over_time

def deduct_time(to_deduct_hr, miss_mins):
    
    if 5 <=  miss_mins< 15 :
        to_deduct_hr += 0.25
    elif 15 <= miss_mins < 30 : 
        to_deduct_hr += 0.50
    elif 30 <= miss_mins < 60 :
        to_deduct_hr += 1.00
    elif miss_mins >= 60:
        #to_deduct_hr += (miss_mins/15+1) * 0.25      
        # more accurate division  13.5 -> 14    
        to_deduct_hr += (miss_mins // 15 + (1 if (miss_mins % 15) >= 7.5 else 0)) * 0.25
    return to_deduct_hr

def read_data():
    # Read the CSV saved from the HTML table
    # Load your data
    df = pd.read_csv('report_data.csv')
    df_len = len(df)
    data = []
    head_cols = df.columns.tolist()
    head_cols_len = len(head_cols)
    weeklyHours = "weeklyHours"
    otFirst3 = "ot first 3"
    otAfter3 = "ot after 3"
    sickLeaveHours = "sickLeaveHours"
    annualLeaveDays = "annualLeaveDays"
    sat = "Sat"
    sun = "Sun"
    comments = "comments"

    for i in range (0,df_len):
        employ_id = int(df.iloc[i, 0])
        if employ_id in (1119,1322):
            continue
        departments = df.iloc[i, 2].split("/")
        position = departments[len(departments)-1]
        employ_week_schedule = {
                
                head_cols[1]: df.iloc[i, 1],
                "emplyee_id": employ_id,
                "emplyee_type" : mapped_dict[employ_id],
                "position" : position,
                #head_cols[3]: df.iloc[i, 3],
                weeklyHours:38.00,
                otFirst3:0,
                otAfter3:0,
                sat:0,
                sun:0,
                sickLeaveHours :0,
                annualLeaveDays:0,
                comments:""
            }

        print("the employee id is ",employ_id)
        if employ_id in [1298]:
            print(employ_id)
        
        carry_over_time_to = to_deduct_hr = to_ot_hr_first_3 =to_ot_hr_second_3 = 0
        weekend_day = sat
        totalSickHours = totalAnnualHours = 0
        comments_list = []
        # remove the previous sunday in index 4
        for j in range(5,head_cols_len-2):
            col_name = head_cols[j]
            print(col_name)
            text = df.iloc[i, j]
            print(text)
            times_list = ast.literal_eval(text)
            active_schedule_diff, come_late_mins, lunch_break, left_early_mins ,carry_over_time = calculate_time(times_list)
            # comes to work in late evening and work until next day on Saturday
            if j == head_cols_len-4 and end_with_a(times_list[3]) and end_with_a(times_list[4]) and end_with_a(times_list[5]):
                carry_over_time = [active_schedule_diff, come_late_mins, lunch_break, left_early_mins]
                active_schedule_diff = come_late_mins = lunch_break = left_early_mins = 0
            
            # comes late in the eveing and working in next day
            if type(carry_over_time_to) == list:
                active_schedule_diff, come_late_mins, lunch_break, left_early_mins = carry_over_time_to[0],carry_over_time_to[1],carry_over_time_to[2],carry_over_time_to[3]
                carry_over_time_to = 0
            # carry the yesterday's time o ver and set it to 0 again
            elif carry_over_time_to != 0 and active_schedule_diff in (None,0):
                    active_schedule_diff = carry_over_time_to
                    carry_over_time_to = 0
            employ_week_schedule[col_name] = [active_schedule_diff, come_late_mins, lunch_break, left_early_mins]
            # if carry_over_time not None, set the carry_over_time_to
            if carry_over_time is not None:
                carry_over_time_to = carry_over_time
                
            
             # these employees already in blacklist. so if over, then deduct hours
            if employ_id in [10024,1162]:
                if lunch_break > 30:
                    to_deduct_hr += 0.25
            else:
                # record lunch break in comments
                if lunch_break > 40:
                    comments_list.append(f"{col_name} long lunch time {lunch_break}")
                    employ_week_schedule[comments] = f"long lunch time {lunch_break}"
            
            # calc officer sick days sickDay 
            if active_schedule_diff == None:
                totalSickHours += 7.6
            else:
                
               # weekend days
                if j in (head_cols_len-4,head_cols_len-3) :
                    # if ot in weekend; active_schedule_diff is the ot hrs
                    if  active_schedule_diff > 0:
                        if j == (head_cols_len-4):
                            weekend_day = sat
                        else:
                            weekend_day = sun
                        # not carry over from friday
                        if active_schedule_diff > 12 :
                            active_schedule_diff = round(active_schedule_diff/60, 2)
                        # CAO, GIA BUU max weekend ot is 6 hrs
                        if employ_id == 10184 and active_schedule_diff > 6:
                                active_schedule_diff = 6.0
                        if active_schedule_diff > 7.6:
                            active_schedule_diff = 7.6
                        employ_week_schedule[weekend_day] = active_schedule_diff
                else:
                    
                    if active_schedule_diff < 0 :
                        to_deduct_hr = deduct_time(to_deduct_hr, come_late_mins)
                        to_deduct_hr = deduct_time(to_deduct_hr, left_early_mins)
                        # most likely , forget to record lunch or left
                        if active_schedule_diff < -120 and (lunch_break == 0 or left_early_mins==0 or come_late_mins ==0):
                            to_deduct_hr += 0.25
                            

                    # non-officer temp worker not working 
                    elif active_schedule_diff == 0 and come_late_mins == 0 and lunch_break == 0 and left_early_mins ==0:
                        to_deduct_hr += 7.6
                    else:
                        # cal ot time  1145 yasser
                        if position in ("Truck Driver",) or employ_id in (1145,9491):
                            # todo 
                            if employ_id == 9491:
                                if come_late_mins > 0 :
                                    continue
                            #  + (1 if (active_schedule_diff % 15) >= 7.5 else 0)
                            to_ot_hr = round((active_schedule_diff// 15) * 0.25,2)
                            if to_ot_hr > 3:
                                to_ot_hr_first_3 += 3
                                to_ot_hr_second_3 += (to_ot_hr-3) 
                            else:
                                to_ot_hr_first_3 += to_ot_hr

 
        # case of annual leave: two sick leave days
        
        if totalSickHours > 7.6*2:
            totalAnnualHours = totalSickHours
            totalSickHours = 0

        if employ_id not in mapped_dict:
            print(f"{employ_id} not in the list")
        else:
            if mapped_dict[employ_id] == official_worker:
                totalAnnualHours += to_deduct_hr
            else:
                employ_week_schedule[weeklyHours] = "{:.2f}".format(38.00 - to_deduct_hr)
        employ_week_schedule[sickLeaveHours] = totalSickHours
        employ_week_schedule[annualLeaveDays] = totalAnnualHours
        #BUU
        if employ_id in (10184,) and to_ot_hr_first_3 > 10:
            to_ot_hr_first_3 = 10
            to_ot_hr_second_3 = 0
        employ_week_schedule[otFirst3] = to_ot_hr_first_3
        employ_week_schedule[otAfter3] = to_ot_hr_second_3
        employ_week_schedule[comments] =  "; ".join(comments_list) 
        data.append(employ_week_schedule)
    df_1 = pd.DataFrame(data)
    df_1 = pd.DataFrame(data).set_index('emplyee_id')  # Assuming 'Employee ID' is the column to match
    df_1 = df_1.reindex(keys_list).reset_index()  # Reset index to make 'Employee ID' a column again
    df_1.to_csv('report_data_process_1.csv', index=False,quoting=csv.QUOTE_NONNUMERIC)
#extra_data()
read_data()

# truck driver no come_late / leave_earli  . calc time_diff for ot and deduct time
# some records only have 6 segements
# 1297  1302    9091