#!/usr/bin/env python3
"""
COSCUP 2025 Data Alignment Script

Takes the new pretalx API data and aligns it with our existing tag classifications,
while keeping the official difficulty levels and updated abstracts.
"""

import json
import re
from typing import Dict, List, Any, Optional

def load_existing_tags_mapping() -> Dict[str, List[str]]:
    """Extract existing tag mappings from embedded_data.go"""
    
    # Read the existing embedded_data.go file
    with open('/home/davidleitw/Desktop/coscup-arranger/mcp/embedded_data.go', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Extract session entries with regex
    session_pattern = r'Code:\s+"([^"]+)".*?Tags:\s+\[\]string\{([^}]+)\}'
    
    tags_mapping = {}
    
    for match in re.finditer(session_pattern, content, re.DOTALL):
        code = match.group(1)
        tags_str = match.group(2)
        
        # Parse tag constants
        tag_constants = re.findall(r'Tag\w+', tags_str)
        
        # Convert tag constants to actual tag strings
        tag_map = {
            'TagAI': '🧠 AI',
            'TagLanguages': '🗣️ Languages',
            'TagWeb3': '⛓️ Web3',
            'TagDatabase': '🗃️ Database',
            'TagSecurity': '🔒 Security',
            'TagHardware': '🛠️ Hardware',
            'TagVehicle': '🚗 Vehicle',
            'TagNetwork': '🌐 Network',
            'TagDevOps': '🚀️ DevOps',
            'TagSystem': '💻 System',
            'TagEnterprise': '🏢 Enterprise',
            'TagData': '📊 Data',
            'TagGaming': '🎮 Gaming',
            'TagAgriculture': '🌾 Agriculture',
            'TagHealthcare': '⚕️ Healthcare',
            'TagKeynote': '🔑 Keynote',
            'TagPolicy': '📜️ Policy',
            'TagGlobal': '🌍 Global',
            'TagOpenData': '👐️ OpenData',
            'TagEducation': '🎓 Education',
            'TagSocial': '🍻 Social',
            'TagSideProject': '💡 SideProject',
        }
        
        tags = [tag_map.get(const, const) for const in tag_constants if const in tag_map]
        if tags:
            tags_mapping[code] = tags
    
    print(f"Loaded {len(tags_mapping)} existing tag mappings")
    return tags_mapping

def generate_aligned_go_data(new_sessions_data: Dict, existing_tags: Dict[str, List[str]]) -> str:
    """Generate Go embedded data with aligned tags and official data"""
    
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

// AlignedCOSCUPData contains COSCUP 2025 session data aligned with existing tags
// Uses official pretalx API data with our existing tag classifications
var AlignedCOSCUPData = map[string]map[string][]Session{
'''
    
    stats = {
        'total_sessions': 0,
        'existing_tags_used': 0,
        'new_tags_generated': 0,
        'missing_sessions': []
    }
    
    # Generate data structure
    for day in sorted(new_sessions_data.keys()):
        go_code += f'\t"{day}": {{\n'
        
        for room in sorted(new_sessions_data[day].keys()):
            sessions = new_sessions_data[day][room]
            go_code += f'\t\t"{room}": {{\n'
            
            for session in sessions:
                stats['total_sessions'] += 1
                
                # Use existing tags if available, otherwise generate new ones
                code = session['code']
                if code in existing_tags:
                    tags = existing_tags[code]
                    stats['existing_tags_used'] += 1
                else:
                    # Generate fallback tags based on track and content
                    tags = generate_fallback_tags(session)
                    stats['new_tags_generated'] += 1
                    stats['missing_sessions'].append(code)
                
                # Format speakers array
                speakers_str = ', '.join(f'"{s}"' for s in session['speakers'])
                
                # Format tags array with constants
                tag_constants = []
                for tag in tags:
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
                
                # Escape strings properly
                title = session['title'].replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')
                track = session['track'].replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')
                abstract = session['abstract'].replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')
                
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
    
    return go_code, stats

def generate_fallback_tags(session: Dict) -> List[str]:
    """Generate fallback tags for sessions not in existing data"""
    tags = []
    
    # Normalize text for comparison
    text = f"{session['track']} {session['title']} {session.get('abstract', '')}".lower()
    
    # Conservative tag classification (only obvious cases)
    if any(keyword in text for keyword in ['security', 'secure', 'attack', 'vulnerability', 'encryption']):
        tags.append("🔒 Security")
        
    if any(keyword in text for keyword in ['kernel', 'linux', 'system', 'os', 'operating']):
        tags.append("💻 System")
        
    if any(keyword in text for keyword in ['ai', 'machine learning', 'neural', 'deep learning', 'llm']):
        tags.append("🧠 AI")
        
    if any(keyword in text for keyword in ['python', 'go', 'rust', 'javascript', 'kotlin', 'java', 'ruby']):
        tags.append("🗣️ Languages")
        
    if any(keyword in text for keyword in ['database', 'sql', 'postgresql', 'mysql']):
        tags.append("🗃️ Database")
        
    if any(keyword in text for keyword in ['devops', 'kubernetes', 'docker', 'cloud']):
        tags.append("🚀️ DevOps")
        
    if any(keyword in text for keyword in ['hardware', 'raspberry pi', 'iot', 'embedded']):
        tags.append("🛠️ Hardware")
        
    if any(keyword in text for keyword in ['blockchain', 'web3', 'cryptocurrency']):
        tags.append("⛓️ Web3")
        
    if any(keyword in text for keyword in ['network', 'tcp', 'http', 'api']):
        tags.append("🌐 Network")
        
    if any(keyword in text for keyword in ['policy', 'license', 'legal', 'governance']):
        tags.append("📜️ Policy")
        
    if any(keyword in text for keyword in ['social', 'community', 'networking', 'chat']):
        tags.append("🍻 Social")
        
    # Default to System if no specific tags found
    if not tags:
        if 'keynote' in text or 'opening' in text:
            tags.append("🔑 Keynote")
        else:
            tags.append("💻 System")  # Conservative default
            
    return tags

def main():
    print("=== COSCUP 2025 Data Alignment Script ===")
    
    # Load new sessions data from pretalx
    print("Loading new pretalx API data...")
    with open('/home/davidleitw/Desktop/coscup-arranger/data/updated_coscup_data.json', 'r', encoding='utf-8') as f:
        new_sessions_data = json.load(f)
    
    # Load existing tag mappings
    print("Loading existing tag mappings...")
    existing_tags = load_existing_tags_mapping()
    
    # Generate aligned Go data
    print("Generating aligned Go data...")
    go_code, stats = generate_aligned_go_data(new_sessions_data, existing_tags)
    
    # Write aligned data
    output_file = '/home/davidleitw/Desktop/coscup-arranger/data/aligned_embedded_data.go'
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(go_code)
    
    print(f"Aligned data written to {output_file}")
    
    # Print statistics
    print("\n=== Statistics ===")
    print(f"Total sessions: {stats['total_sessions']}")
    print(f"Sessions using existing tags: {stats['existing_tags_used']}")
    print(f"Sessions with generated tags: {stats['new_tags_generated']}")
    
    if stats['missing_sessions']:
        print(f"\nSessions with generated tags (first 10): {stats['missing_sessions'][:10]}")
        if len(stats['missing_sessions']) > 10:
            print(f"... and {len(stats['missing_sessions']) - 10} more")
    
    # Verify key sessions
    print("\n=== Key Session Verification ===")
    
    # Check JPADKC specifically
    jpadkc_found = False
    for day in new_sessions_data:
        for room in new_sessions_data[day]:
            for session in new_sessions_data[day][room]:
                if session['code'] == 'JPADKC':
                    print(f"JPADKC session verified:")
                    print(f"  Time: {session['start']}-{session['end']}")
                    print(f"  Room: {session['room']}")
                    print(f"  Difficulty: {session['difficulty']} (official)")
                    
                    # Show which tags were used
                    if session['code'] in existing_tags:
                        print(f"  Tags: {existing_tags[session['code']]} (existing)")
                    else:
                        fallback_tags = generate_fallback_tags(session)
                        print(f"  Tags: {fallback_tags} (generated)")
                    
                    jpadkc_found = True
                    break
    
    if not jpadkc_found:
        print("WARNING: JPADKC session not found!")

if __name__ == '__main__':
    main()