import os
import re

templates_dir = '/home/rdeavila/dev/txlog/server/templates'
phosphor_raw_dir = '/tmp/phosphor-icons/raw'

def extract_paths(svg_content):
    # Extract all 'd' attributes from <path> tags
    paths = set(re.findall(r'<path[^>]*d="([^"]+)"', svg_content))
    # Also extract points from <polygon> or <polyline> just in case, though usually paths are used
    points = set(re.findall(r'points="([^"]+)"', svg_content))
    return frozenset(paths | points)

# Build a map of path signatures to regular SVG content
fill_dir = os.path.join(phosphor_raw_dir, 'fill')
regular_dir = os.path.join(phosphor_raw_dir, 'regular')

sig_to_regular_svg = {}
icon_names = set()

for filename in os.listdir(fill_dir):
    if not filename.endswith('.svg'):
        continue
    
    # E.g. "house-fill.svg" -> "house"
    base_name = filename.replace('-fill.svg', '')
    regular_filename = f"{base_name}.svg"
    regular_path = os.path.join(regular_dir, regular_filename)
    
    if not os.path.exists(regular_path):
        continue
        
    with open(os.path.join(fill_dir, filename), 'r') as f:
        fill_svg = f.read()
    
    with open(regular_path, 'r') as f:
        regular_svg = f.read()
        
    sig = extract_paths(fill_svg)
    if sig:
        sig_to_regular_svg[sig] = regular_svg

# Now process all HTML templates
svg_tag_pattern = re.compile(r'(<svg[^>]*>)(.*?)</svg>', re.DOTALL)
class_pattern = re.compile(r'class="([^"]+)"')

changed_files = 0
replaced_count = 0

for root_dir, _, files in os.walk(templates_dir):
    for filename in files:
        if not filename.endswith('.html'):
            continue
            
        filepath = os.path.join(root_dir, filename)
        with open(filepath, 'r') as f:
            content = f.read()
            
        original_content = content
        
        # We need to replace SVGs in the content
        def replace_svg(match):
            global replaced_count
            svg_open_tag = match.group(1)
            inner_content = match.group(2)
            
            sig = extract_paths(inner_content)
            if not sig:
                return match.group(0) # no paths, skip
                
            # Check if this signature matches a Phosphor fill icon
            if sig in sig_to_regular_svg:
                regular_svg_content = sig_to_regular_svg[sig]
                
                # We need to inject the classes from the original SVG into the regular SVG
                # Extract classes from original
                cls_match = class_pattern.search(svg_open_tag)
                classes = cls_match.group(1) if cls_match else ""
                
                # Find the open tag of the regular SVG
                reg_open_tag_match = re.search(r'<svg[^>]*>', regular_svg_content)
                if not reg_open_tag_match:
                    return match.group(0)
                    
                reg_open_tag = reg_open_tag_match.group(0)
                reg_inner = regular_svg_content[reg_open_tag_match.end():-6] # strip <svg> and </svg>
                
                # We want to keep original attributes like width/height/viewBox if possible,
                # but Phosphor regular has its own viewBox="0 0 256 256". 
                # Let's just use the original svg_open_tag but remove fill="currentColor" if present
                # Actually, regular icons need fill="none" and stroke="currentColor" stroke-width="16" etc on their paths.
                # Since Phosphor puts those on the individual paths/rects, we can just use the original open tag!
                # Wait, no. Phosphor's downloaded SVGs have fill="none" stroke="currentColor" on the paths.
                # So we can just use the original svg open tag, but remove `fill="currentColor"` from it.
                
                new_open_tag = re.sub(r'\bfill="currentColor"', '', svg_open_tag)
                new_open_tag = re.sub(r'\bfill=\'currentColor\'', '', new_open_tag)
                # also remove any multiple spaces
                new_open_tag = re.sub(r'\s+', ' ', new_open_tag).replace(' >', '>')
                
                replaced_count += 1
                return f"{new_open_tag}{reg_inner}</svg>"
                
            return match.group(0)
            
        new_content = svg_tag_pattern.sub(replace_svg, content)
        
        if new_content != original_content:
            with open(filepath, 'w') as f:
                f.write(new_content)
            changed_files += 1

print(f"Replaced {replaced_count} icons across {changed_files} files.")
