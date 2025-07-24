#!/usr/bin/env python3
import requests
import json
import re
from datetime import datetime
import ast

def fetch_sessions_from_js():
    """從 JavaScript 資源文件中獲取議程資料"""
    
    js_url = "https://coscup.org/2025/assets/chunks/allSubmissions.zh-tw.data.BUNdBk1a.js"
    
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
    }
    
    response = requests.get(js_url, headers=headers)
    response.raise_for_status()
    
    js_content = response.text
    
    # 尋找 JSON 資料
    # 格式：const e=JSON.parse(`[...]`)
    json_match = re.search(r'JSON\.parse\(`(\[.*?\])`\)', js_content, re.DOTALL)
    if not json_match:
        # 嘗試其他可能的格式
        json_match = re.search(r'=\s*JSON\.parse\(`(\[.*\])`\)', js_content, re.DOTALL)
    
    if json_match:
        json_str = json_match.group(1)
        
        # 嘗試使用 eval 方式解析（安全的，因為我們知道這是 JSON 資料）
        try:
            # 將 JavaScript 字串轉換為 Python 可解析的格式
            # 處理常見的轉義序列
            json_str = json_str.replace('\\`', '`')  # 處理轉義的反引號
            json_str = json_str.replace('\\"', '"')   # 處理轉義的雙引號
            json_str = json_str.replace('\\\\', '\\') # 處理雙反斜線
            
            sessions_data = json.loads(json_str)
            return sessions_data
            
        except json.JSONDecodeError as e:
            print(f"JSON 解析錯誤: {e}")
            
            # 嘗試更深入的修復
            try:
                # 移除可能有問題的轉義序列
                fixed_json = re.sub(r'\\([^"\\nt])', r'\1', json_str)
                sessions_data = json.loads(fixed_json)
                return sessions_data
            except:
                print(f"深度修復也失敗，嘗試其他方法...")
                
                # 最後嘗試：分段解析
                try:
                    # 提取前100個字符查看格式
                    preview = json_str[:500]
                    print(f"JSON 開頭內容預覽: {preview}")
                    return None
                except Exception as inner_e:
                    print(f"所有解析方法都失敗: {inner_e}")
                    return None
    else:
        print("無法在 JS 文件中找到 JSON 資料")
        return None

def filter_sessions_by_room_and_date(sessions, room_name, target_date):
    """根據會議室和日期篩選議程"""
    
    filtered_sessions = []
    
    for session in sessions:
        session_room = session.get('room', {}).get('name', '') if session.get('room') else ''
        session_start = session.get('start', '')
        
        # 檢查日期
        if target_date in session_start and room_name in session_room:
            filtered_sessions.append({
                'title': session.get('title', ''),
                'speaker': ', '.join([speaker.get('name', '') for speaker in session.get('speakers', [])]),
                'time': f"{session.get('start', '').split('T')[1][:5]} ~ {session.get('end', '').split('T')[1][:5]}" if session.get('start') and session.get('end') else '',
                'room': session_room,
                'track': session.get('track', {}).get('name', '') if session.get('track') else '',
                'abstract': session.get('abstract', ''),
                'start': session.get('start', ''),
                'end': session.get('end', '')
            })
    
    return sorted(filtered_sessions, key=lambda x: x['start'])

def save_sessions_data(sessions, filename='coscup_2025_from_js.json'):
    """保存議程資料"""
    data = {
        'conference': 'COSCUP x RubyConf Taiwan 2025',
        'source': 'JavaScript Resource File',
        'scraped_at': datetime.now().isoformat(),
        'total_sessions': len(sessions),
        'sessions': sessions
    }
    
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    
    print(f"資料已保存到 {filename}")

if __name__ == "__main__":
    print("從 JavaScript 資源獲取 COSCUP 2025 議程資料...")
    
    # 獲取完整議程資料
    all_sessions = fetch_sessions_from_js()
    
    if all_sessions:
        print(f"成功獲取 {len(all_sessions)} 個議程項目")
        
        # 保存完整資料
        save_sessions_data(all_sessions)
        
        # 查找 Aug.9 RB101 的議程
        rb101_aug9 = filter_sessions_by_room_and_date(all_sessions, 'RB101', '2025-08-09')
        
        print(f"\n=== Aug.9 RB101 議程 ===")
        if rb101_aug9:
            for i, session in enumerate(rb101_aug9, 1):
                print(f"\n{i}. {session['title']}")
                print(f"   時間: {session['time']}")
                print(f"   講者: {session['speaker']}")
                print(f"   軌道: {session['track']}")
        else:
            print("Aug.9 RB101 沒有安排議程")
        
        # 顯示所有可用的會議室
        all_rooms = set()
        for session in all_sessions:
            room = session.get('room', {}).get('name', '') if session.get('room') else ''
            if room:
                all_rooms.add(room)
        
        print(f"\n=== 所有會議室列表 ===")
        for room in sorted(all_rooms):
            print(f"- {room}")
            
    else:
        print("無法獲取議程資料")