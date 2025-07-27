#!/usr/bin/env python3
"""
COSCUP 2025 Pretalx API Data Fetcher

Fetches the latest session data from pretalx API and generates Go embedded data format.
"""

import requests
import json
from datetime import datetime
from typing import Dict, List, Any, Optional
import argparse

class PretalxDataFetcher:
    def __init__(self, year: int = 2025):
        self.year = year
        self.base_url = "https://pretalx.coscup.org"
        self.event = f"coscup-{year}"
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': f'coscup-data-fetcher/{year}',
            'Accept': 'application/json'
        })
        
    def fetch_paginated(self, endpoint: str, params: Dict = None) -> List[Dict]:
        """Fetch all pages from a paginated API endpoint"""
        all_results = []
        url = f"{self.base_url}/api/events/{self.event}/{endpoint}/"
        
        while url:
            print(f"Fetching: {url}")
            response = self.session.get(url, params=params if url == f"{self.base_url}/api/events/{self.event}/{endpoint}/" else None)
            response.raise_for_status()
            
            data = response.json()
            all_results.extend(data.get('results', []))
            url = data.get('next')
            params = None  # Clear params for subsequent requests
            
        return all_results
    
    def fetch_submissions(self) -> List[Dict]:
        """Fetch all confirmed submissions with expanded data"""
        params = {
            'state': 'confirmed',
            'expand': 'answers,slots',
            'page_size': 50
        }
        return self.fetch_paginated('submissions', params)
    
    def fetch_speakers(self) -> Dict[str, Dict]:
        """Fetch all speakers and return as code -> speaker mapping"""
        speakers = self.fetch_paginated('speakers')
        return {speaker['code']: speaker for speaker in speakers}
    
    def fetch_rooms(self) -> Dict[int, Dict]:
        """Fetch all rooms and return as id -> room mapping"""
        rooms = self.fetch_paginated('rooms')
        return {room['id']: room for room in rooms}
    
    def fetch_tracks(self) -> Dict[int, Dict]:
        """Fetch all tracks and return as id -> track mapping"""
        tracks = self.fetch_paginated('tracks')
        return {track['id']: track for track in tracks}
        
    def parse_slot_time(self, slot: Dict) -> tuple[str, str, str]:
        """Parse slot time and return start, end, day"""
        start_dt = datetime.fromisoformat(slot['start'].replace('Z', '+00:00'))
        end_dt = datetime.fromisoformat(slot['end'].replace('Z', '+00:00'))
        
        # Convert to Taiwan time (UTC+8)
        taiwan_start = start_dt.replace(tzinfo=None)
        taiwan_end = end_dt.replace(tzinfo=None)
        
        day = f"Aug.{taiwan_start.day}"
        start_time = taiwan_start.strftime("%H:%M")
        end_time = taiwan_end.strftime("%H:%M")
        
        return start_time, end_time, day
    
    def get_answer_value(self, answers: List[Dict], question_id: int) -> Optional[str]:
        """Extract answer value for a specific question ID"""
        for answer in answers:
            if answer.get('question') == question_id:
                return answer.get('answer', '').strip()
        return None
    
    def classify_session_tags(self, track_name: str, title: str, abstract: str) -> List[str]:
        """Classify session into appropriate tags based on content"""
        tags = []
        
        # Normalize text for comparison
        text = f"{track_name} {title} {abstract}".lower()
        
        # Tag classification rules
        if any(keyword in text for keyword in ['ai', 'machine learning', 'neural', 'deep learning', 'llm', 'chatgpt']):
            tags.append("🧠 AI")
        
        if any(keyword in text for keyword in ['security', 'secure', 'attack', 'vulnerability', 'encryption', 'privacy']):
            tags.append("🔒 Security")
            
        if any(keyword in text for keyword in ['python', 'go', 'rust', 'javascript', 'kotlin', 'java', 'ruby']):
            tags.append("🗣️ Languages")
            
        if any(keyword in text for keyword in ['database', 'sql', 'postgresql', 'mysql', 'redis']):
            tags.append("🗃️ Database")
            
        if any(keyword in text for keyword in ['devops', 'kubernetes', 'docker', 'cloud', 'deployment']):
            tags.append("🚀️ DevOps")
            
        if any(keyword in text for keyword in ['hardware', 'raspberry pi', 'iot', 'embedded']):
            tags.append("🛠️ Hardware")
            
        if any(keyword in text for keyword in ['blockchain', 'web3', 'cryptocurrency', 'nft']):
            tags.append("⛓️ Web3")
            
        if any(keyword in text for keyword in ['network', 'tcp', 'http', 'api']):
            tags.append("🌐 Network")
            
        if any(keyword in text for keyword in ['system', 'linux', 'kernel', 'operating']):
            tags.append("💻 System")
            
        if any(keyword in text for keyword in ['education', 'teaching', 'learning', 'student']):
            tags.append("🎓 Education")
            
        if any(keyword in text for keyword in ['policy', 'license', 'legal', 'governance']):
            tags.append("📜️ Policy")
            
        if any(keyword in text for keyword in ['social', 'community', 'networking', 'chat']):
            tags.append("🍻 Social")
            
        # Default to general category if no specific tags
        if not tags:
            if 'keynote' in text or 'opening' in text:
                tags.append("🔑 Keynote")
            else:
                tags.append("💻 System")  # Default category
                
        return tags
    
    def process_session(self, submission: Dict, speakers: Dict, rooms: Dict, tracks: Dict) -> Optional[Dict]:
        """Process a single submission into session format"""
        
        # Skip if no slots
        if not submission.get('slots'):
            print(f"Skipping {submission['code']}: No slots")
            return None
            
        slot = submission['slots'][0]
        if not slot.get('start') or not slot.get('end') or not slot.get('room'):
            print(f"Skipping {submission['code']}: Incomplete slot data")
            return None
            
        # Parse time and room
        start_time, end_time, day = self.parse_slot_time(slot)
        room_id = slot['room']
        room_info = rooms.get(room_id, {})
        room_name = room_info.get('name', {}).get('en') or room_info.get('name', {}).get('zh-tw') or f"Room{room_id}"
        
        # Get track info
        track_id = submission.get('track')
        track_info = tracks.get(track_id, {}) if track_id else {}
        track_name = track_info.get('name', {}).get('zh-tw') or track_info.get('name', {}).get('en') or "General"
        
        # Get speaker names
        speaker_names = []
        for speaker_code in submission.get('speakers', []):
            speaker = speakers.get(speaker_code, {})
            speaker_names.append(speaker.get('name', speaker_code))
        
        # Get answers for additional metadata
        answers = submission.get('answers', [])
        
        # Question IDs based on COSCUP submission form
        difficulty = self.get_answer_value(answers, 59) or "入門"  # Default difficulty
        language = self.get_answer_value(answers, 57) or "漢語"   # Default language
        
        # Map difficulty values
        difficulty_map = {
            'Beginner': '入門',
            'Intermediate': '中階', 
            'Advanced': '進階'
        }
        difficulty = difficulty_map.get(difficulty, difficulty)
        
        # Map language values
        language_map = {
            'Chinese': '漢語',
            'English': '英語',
            'Japanese': '日本語'
        }
        language = language_map.get(language, language)
        
        # Generate tags
        abstract = submission.get('abstract', '') or ''
        tags = self.classify_session_tags(track_name, submission['title'], abstract)
        
        return {
            'code': submission['code'],
            'title': submission['title'],
            'speakers': speaker_names,
            'start': start_time,
            'end': end_time,
            'track': track_name,
            'abstract': abstract[:200] + '...' if len(abstract) > 200 else abstract,  # Truncate for embedded data
            'language': language,
            'difficulty': difficulty,
            'room': room_name,
            'day': day,
            'tags': tags,
            'url': f"https://coscup.org/2025/sessions/{submission['code']}"
        }
    
    def generate_go_data(self, sessions_by_day_room: Dict) -> str:
        """Generate Go embedded data format"""
        
        # Start with package and constants
        go_code = '''package mcp

// Universal session tags with emojis
const (
	TagAI          = "🧠 AI"
	TagLanguages   = "🗣️ Languages"
	TagWeb3        = "⛓️ Web3"
	TagDatabase    = "🗃️ Database"
	TagSecurity    = "🔒 Security"
	TagHardware    = "🛠️ Hardware"
	TagVehicle     = "🚗 Vehicle"
	TagNetwork     = "🌐 Network"
	TagDevOps      = "🚀️ DevOps"
	TagSystem      = "💻 System"
	TagEnterprise  = "🏢 Enterprise"
	TagData        = "📊 Data"
	TagGaming      = "🎮 Gaming"
	TagAgriculture = "🌾 Agriculture"
	TagHealthcare  = "⚕️ Healthcare"
	TagKeynote     = "🔑 Keynote"
	TagPolicy      = "📜️ Policy"
	TagGlobal      = "🌍 Global"
	TagOpenData    = "👐️ OpenData"
	TagEducation   = "🎓 Education"
	TagSocial      = "🍻 Social"
	TagSideProject = "💡 SideProject"
)

// UpdatedCOSCUPData contains COSCUP 2025 session data from pretalx API
// Generated from pretalx API - ''' + datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '''
var UpdatedCOSCUPData = map[string]map[string][]Session{
'''
        
        # Generate data structure
        for day in sorted(sessions_by_day_room.keys()):
            go_code += f'\t"{day}": {{\n'
            
            for room in sorted(sessions_by_day_room[day].keys()):
                sessions = sessions_by_day_room[day][room]
                go_code += f'\t\t"{room}": {{\n'
                
                for session in sessions:
                    # Format speakers array
                    speakers_str = ', '.join(f'"{s}"' for s in session['speakers'])
                    
                    # Format tags array
                    tag_constants = []
                    for tag in session['tags']:
                        tag_map = {
                            "🧠 AI": "TagAI",
                            "🗣️ Languages": "TagLanguages", 
                            "⛓️ Web3": "TagWeb3",
                            "🗃️ Database": "TagDatabase",
                            "🔒 Security": "TagSecurity",
                            "🛠️ Hardware": "TagHardware",
                            "🚗 Vehicle": "TagVehicle",
                            "🌐 Network": "TagNetwork",
                            "🚀️ DevOps": "TagDevOps",
                            "💻 System": "TagSystem",
                            "🏢 Enterprise": "TagEnterprise",
                            "📊 Data": "TagData",
                            "🎮 Gaming": "TagGaming",
                            "🌾 Agriculture": "TagAgriculture",
                            "⚕️ Healthcare": "TagHealthcare",
                            "🔑 Keynote": "TagKeynote",
                            "📜️ Policy": "TagPolicy",
                            "🌍 Global": "TagGlobal",
                            "👐️ OpenData": "TagOpenData",
                            "🎓 Education": "TagEducation",
                            "🍻 Social": "TagSocial",
                            "💡 SideProject": "TagSideProject",
                        }
                        tag_constants.append(tag_map.get(tag, f'"{tag}"'))
                    
                    tags_str = ', '.join(tag_constants)
                    
                    # Escape strings
                    title = session['title'].replace('"', '\\"').replace('\n', '\\n')
                    track = session['track'].replace('"', '\\"').replace('\n', '\\n')
                    abstract = session['abstract'].replace('"', '\\"').replace('\n', '\\n')
                    
                    go_code += f'''\t\t\t{{
\t\t\t\tCode:       "{session['code']}",
\t\t\t\tTitle:      "{title}",
\t\t\t\tSpeakers:   []string{{{speakers_str}}},
\t\t\t\tStart:      "{session['start']}",
\t\t\t\tEnd:        "{session['end']}",
\t\t\t\tTrack:      "{track}",
\t\t\t\tAbstract:   "{abstract}",
\t\t\t\tLanguage:   "{session['language']}",
\t\t\t\tDifficulty: "{session['difficulty']}",
\t\t\t\tRoom:       "{session['room']}",
\t\t\t\tDay:        "{session['day']}",
\t\t\t\tURL:        "{session['url']}",
\t\t\t\tTags:       []string{{{tags_str}}},
\t\t\t}},
'''
                
                go_code += '\t\t},\n'
            
            go_code += '\t},\n'
        
        go_code += '}\n'
        
        return go_code
    
    def run(self) -> Dict:
        """Main execution method"""
        print(f"Fetching COSCUP {self.year} data from pretalx API...")
        
        # Fetch all required data
        print("Fetching submissions...")
        submissions = self.fetch_submissions()
        print(f"Found {len(submissions)} submissions")
        
        print("Fetching speakers...")
        speakers = self.fetch_speakers()
        print(f"Found {len(speakers)} speakers")
        
        print("Fetching rooms...")
        rooms = self.fetch_rooms()
        print(f"Found {len(rooms)} rooms")
        
        print("Fetching tracks...")
        tracks = self.fetch_tracks()
        print(f"Found {len(tracks)} tracks")
        
        # Process sessions
        print("Processing sessions...")
        sessions_by_day_room = {}
        processed_count = 0
        
        for submission in submissions:
            session = self.process_session(submission, speakers, rooms, tracks)
            if session:
                day = session['day']
                room = session['room']
                
                if day not in sessions_by_day_room:
                    sessions_by_day_room[day] = {}
                if room not in sessions_by_day_room[day]:
                    sessions_by_day_room[day][room] = []
                    
                sessions_by_day_room[day][room].append(session)
                processed_count += 1
        
        # Sort sessions by start time within each room
        for day in sessions_by_day_room:
            for room in sessions_by_day_room[day]:
                sessions_by_day_room[day][room].sort(key=lambda s: s['start'])
        
        print(f"Processed {processed_count} sessions")
        print(f"Found sessions for days: {list(sessions_by_day_room.keys())}")
        
        return sessions_by_day_room

def main():
    parser = argparse.ArgumentParser(description='Fetch COSCUP data from pretalx API')
    parser.add_argument('--year', type=int, default=2025, help='COSCUP year')
    parser.add_argument('--output-json', type=str, help='Output JSON file path')
    parser.add_argument('--output-go', type=str, help='Output Go file path')
    
    args = parser.parse_args()
    
    fetcher = PretalxDataFetcher(args.year)
    
    try:
        sessions_data = fetcher.run()
        
        # Output JSON if requested
        if args.output_json:
            with open(args.output_json, 'w', encoding='utf-8') as f:
                json.dump(sessions_data, f, indent=2, ensure_ascii=False)
            print(f"JSON data written to {args.output_json}")
        
        # Output Go if requested  
        if args.output_go:
            go_code = fetcher.generate_go_data(sessions_data)
            with open(args.output_go, 'w', encoding='utf-8') as f:
                f.write(go_code)
            print(f"Go data written to {args.output_go}")
        
        # Verify doraeric session
        print("\n=== Verification ===")
        found_jpadkc = False
        for day in sessions_data:
            for room in sessions_data[day]:
                for session in sessions_data[day][room]:
                    if session['code'] == 'JPADKC':
                        print(f"Found JPADKC session:")
                        print(f"  Title: {session['title']}")
                        print(f"  Speaker: {session['speakers']}")
                        print(f"  Time: {session['start']}-{session['end']}")
                        print(f"  Room: {session['room']}")
                        found_jpadkc = True
                        break
        
        if not found_jpadkc:
            print("WARNING: JPADKC session not found!")
            
    except Exception as e:
        print(f"Error: {e}")
        return 1
    
    return 0

if __name__ == '__main__':
    exit(main())