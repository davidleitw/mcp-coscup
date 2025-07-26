#!/usr/bin/env python3
"""
Generate embedded_data.go from JSON with proper tags
"""

import json
import re

# Tag mapping based on track and content
TAG_MAPPING = {
    "AI": "🧠 AI",
    "Languages": "🗣️ Languages", 
    "Web3": "⛓️ Web3",
    "Database": "🗃️ Database",
    "Security": "🔒 Security",
    "Hardware": "🛠️ Hardware",
    "Vehicle": "🚗 Vehicle",
    "Network": "🌐 Network",
    "DevOps": "🚀️ DevOps",
    "System": "💻 System",
    "Enterprise": "🏢 Enterprise",
    "Data": "📊 Data",
    "Gaming": "🎮 Gaming",
    "Agriculture": "🌾 Agriculture",
    "Healthcare": "⚕️ Healthcare",
    "Keynote": "🔑 Keynote",
    "Policy": "📜️ Policy",
    "Global": "🌍 Global",
    "OpenData": "👐️ OpenData",
    "Education": "🎓 Education",
    "Social": "🍻 Social",
    "SideProject": "💡 SideProject"
}

def generate_tags(session):
    """Generate tags for a session based on track, title and abstract"""
    tags = []
    
    track = session.get('track', '').lower()
    title = session.get('title', '').lower()
    abstract = session.get('abstract', '').lower()
    
    # AI & Machine Learning
    if any(keyword in track for keyword in ['ai', 'machine learning']) or \
       any(keyword in title for keyword in ['ai', 'ml', '機器學習', '人工智慧', 'llm', 'agent']) or \
       any(keyword in abstract for keyword in ['ai', 'machine learning', 'neural', 'model']):
        tags.append('TagAI')
    
    # Programming Languages
    if any(keyword in track for keyword in ['golang', 'python', 'ruby', 'kotlin', 'swift', 'jvm', 'javascript']) or \
       any(keyword in title for keyword in ['go', 'python', 'ruby', 'kotlin', 'swift', 'java', 'js', 'programming']):
        tags.append('TagLanguages')
    
    # Security & Privacy
    if any(keyword in track for keyword in ['security', 'tor', 'hitcon']) or \
       any(keyword in title for keyword in ['security', 'hack', '隱私', '安全', 'attack', 'privacy']):
        tags.append('TagSecurity')
    
    # Web3 & Blockchain
    if any(keyword in track for keyword in ['blockchain', 'web3']) or \
       any(keyword in title for keyword in ['blockchain', 'web3', 'crypto', 'defi', '區塊鏈']):
        tags.append('TagWeb3')
    
    # Database
    if any(keyword in track for keyword in ['postgresql', 'database']) or \
       any(keyword in title for keyword in ['database', '資料庫', 'sql', 'postgresql']):
        tags.append('TagDatabase')
    
    # Hardware & System
    if any(keyword in track for keyword in ['hardware', 'firmware', 'system']) or \
       any(keyword in title for keyword in ['hardware', 'firmware', 'risc-v', 'fpga', 'embedded']):
        tags.append('TagHardware')
    
    # Vehicle
    if any(keyword in track for keyword in ['vehicle', 'automotive']) or \
       any(keyword in title for keyword in ['vehicle', 'automotive', 'sdv']):
        tags.append('TagVehicle')
    
    # Network
    if any(keyword in track for keyword in ['network', 'networking']) or \
       any(keyword in title for keyword in ['network', '網路', 'tcp', 'http', 'api']):
        tags.append('TagNetwork')
    
    # DevOps
    if any(keyword in track for keyword in ['devops', 'cloud', 'monitoring']) or \
       any(keyword in title for keyword in ['devops', 'docker', 'kubernetes', 'ci/cd', 'monitoring']):
        tags.append('TagDevOps')
    
    # System Software
    if any(keyword in track for keyword in ['system software']) or \
       any(keyword in title for keyword in ['kernel', 'linux', 'system', '系統']):
        tags.append('TagSystem')
    
    # Enterprise
    if any(keyword in track for keyword in ['enterprise', 'odoo', 'erp']) or \
       any(keyword in title for keyword in ['enterprise', 'erp', 'business']):
        tags.append('TagEnterprise')
    
    # Data & Analytics
    if any(keyword in title for keyword in ['data', '資料', 'analytics', 'visualization', '分析']):
        tags.append('TagData')
    
    # Policy & Licensing
    if any(keyword in track for keyword in ['policy', 'licensing']) or \
       any(keyword in title for keyword in ['policy', 'license', '政策', '授權', '開源', 'legal']):
        tags.append('TagPolicy')
    
    # Global & International
    if any(keyword in track for keyword in ['japan', 'global', 'world tour', 'international']) or \
       any(keyword in title for keyword in ['fosdem', 'japan', '日本', 'international', 'global']):
        tags.append('TagGlobal')
    
    # OpenData
    if any(keyword in track for keyword in ['openstreet', 'wikidata', 'wikipedia']) or \
       any(keyword in title for keyword in ['openstreet', 'wikidata', 'wikipedia', '開放資料']):
        tags.append('TagOpenData')
    
    # Education
    if any(keyword in title for keyword in ['教學', '入門', 'beginner', 'tutorial', 'education', '學習']):
        tags.append('TagEducation')
    
    # Social & Networking
    if any(keyword in track for keyword in ['unconference']) or \
       any(keyword in title for keyword in ['bof', 'hacking corner', 'social', '交流']):
        tags.append('TagSocial')
    
    # Side Projects
    if any(keyword in track for keyword in ['side project']) or \
       any(keyword in title for keyword in ['side project', 'indie', 'startup']):
        tags.append('TagSideProject')
    
    # Main Session / Keynote
    if any(keyword in track for keyword in ['main session']) or \
       any(keyword in title for keyword in ['welcome', 'keynote', 'closing']):
        tags.append('TagKeynote')
    
    # Default tag if no specific category
    if not tags:
        tags.append('TagSystem')  # Default to system/tech
    
    return tags[:2]  # Limit to 2 tags max

def escape_go_string(s):
    """Escape special characters for Go string literals"""
    if s is None:
        return '""'
    # Escape backslashes, quotes, and control characters
    s = s.replace('\\', '\\\\')
    s = s.replace('"', '\\"')
    s = s.replace('\r', '\\r')
    s = s.replace('\n', '\\n')
    s = s.replace('\t', '\\t')
    return f'"{s}"'

def generate_go_file():
    """Generate the complete embedded_data.go file"""
    
    # Read JSON data
    with open('data/coscup_2025_by_day_room.json', 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Start building the Go file
    go_content = '''package mcp

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

// EmbeddedCOSCUPData contains hardcoded COSCUP 2025 session data
// Generated from coscup_2025_by_day_room.json
var EmbeddedCOSCUPData = map[string]map[string][]Session{
'''
    
    # Process each day
    for day, rooms in data['structure'].items():
        go_content += f'\t"{day}": {{\n'
        
        # Process each room
        for room, sessions in rooms.items():
            go_content += f'\t\t"{room}": {{\n'
            
            # Process each session
            for session in sessions:
                tags = generate_tags(session)
                tags_str = ', '.join(tags)
                
                speakers_list = ', '.join([escape_go_string(speaker) for speaker in session.get('speakers', [])])
                
                go_content += f'''\t\t\t{{
\t\t\t\tCode:       {escape_go_string(session.get('code', ''))},
\t\t\t\tTitle:      {escape_go_string(session.get('title', ''))},
\t\t\t\tSpeakers:   []string{{{speakers_list}}},
\t\t\t\tStart:      {escape_go_string(session.get('time', {}).get('start', ''))},
\t\t\t\tEnd:        {escape_go_string(session.get('time', {}).get('end', ''))},
\t\t\t\tTrack:      {escape_go_string(session.get('track', ''))},
\t\t\t\tAbstract:   {escape_go_string(session.get('abstract', ''))},
\t\t\t\tLanguage:   {escape_go_string(session.get('language', ''))},
\t\t\t\tDifficulty: {escape_go_string(session.get('difficulty', ''))},
\t\t\t\tRoom:       "{room}",
\t\t\t\tDay:        "{day}",
\t\t\t\tTags:       []string{{{tags_str}}},
\t\t\t}},
'''
            
            go_content += '\t\t},\n'
        
        go_content += '\t},\n'
    
    go_content += '}\n'
    
    return go_content

if __name__ == '__main__':
    print("Generating embedded_data.go...")
    go_code = generate_go_file()
    
    # Write to file
    with open('mcp/embedded_data.go', 'w', encoding='utf-8') as f:
        f.write(go_code)
    
    print("✅ embedded_data.go generated successfully!")
    
    # Count sessions for verification
    with open('data/coscup_2025_by_day_room.json', 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    total_sessions = 0
    for day, rooms in data['structure'].items():
        for room, sessions in rooms.items():
            total_sessions += len(sessions)
            
    print(f"📊 Total sessions processed: {total_sessions}")