import os
import re
import sqlite3
import sys
import urllib.parse
from pathlib import Path

import requests

API_KEY = os.environ["GOOGLE_MAPS_API_KEY"]

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DB_PATH = REPO_ROOT / "data.sqlite3"
OUTPUT_DIR = REPO_ROOT / "static" / "images" / "maps"
ZOOM = 3
SIZE = "300x300"


def extract_location_from_text(text):
    """Try to extract a specific place name from free text.

    Looks for patterns like:
      - "occurred in Lourdes, France"
      - "reported in Soufanieh, a suburb of Damascus"
      - "at the Hill of Tepeyac"
      - decimal coordinates like "46.1234, 5.6789"
    Returns a geocodable string or None.
    """
    if not text:
        return None

    # Decimal coordinates
    coord = re.search(r"(-?\d{1,3}\.\d+)\s*,\s*(-?\d{1,3}\.\d+)", text)
    if coord:
        lat, lng = float(coord.group(1)), float(coord.group(2))
        if -90 <= lat <= 90 and -180 <= lng <= 180:
            return f"{lat},{lng}"

    # "in <Place>, <Region/Country>" — first sentence or two
    snippet = text[:600]
    pattern = re.search(
        r"\bin\s+([A-Z][A-Za-zÀ-ÿ'\-]+(?:\s+[A-Z][A-Za-zÀ-ÿ'\-]+)*)"
        r"(?:,\s*([A-Z][A-Za-zÀ-ÿ'\-]+(?:\s+[A-Z][A-Za-zÀ-ÿ'\-]+)*))?",
        snippet,
    )
    if pattern:
        place = pattern.group(1)
        region = pattern.group(2)
        if region:
            return f"{place}, {region}"
        return place

    return None


def best_center(name, country, description, block_texts):
    """Return the best geocodable center string for this event."""
    for text in [description] + block_texts:
        loc = extract_location_from_text(text)
        if loc:
            return loc

    # Fall back to name + country
    parts = [name, country] if country else [name]
    return ", ".join(p.strip() for p in parts if p)


def fetch_map(center, output_path):
    encoded = urllib.parse.quote(center)
    url = (
        "https://maps.googleapis.com/maps/api/staticmap"
        f"?center={encoded}"
        f"&zoom={ZOOM}"
        f"&size={SIZE}"
        "&maptype=roadmap"
        f"&markers=color:red%7Csize:small%7C{encoded}"
        f"&key={API_KEY}"
    )
    response = requests.get(url, timeout=15)
    response.raise_for_status()
    with open(output_path, "wb") as f:
        f.write(response.content)


def main():
    force = "--force" in sys.argv

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    events = cur.execute(
        "SELECT id, slug, name, country, description FROM events"
        " WHERE slug IS NOT NULL AND slug != ''"
    ).fetchall()

    print(f"Found {len(events)} events with slugs.")

    for event in events:
        slug = event["slug"]
        output_path = OUTPUT_DIR / f"{slug}.png"

        if output_path.exists() and not force:
            print(f"  skip  {slug}  (already exists, use --force to overwrite)")
            continue

        blocks = cur.execute(
            "SELECT content FROM event_blocks"
            " WHERE event_id = ? AND content IS NOT NULL",
            (event["id"],),
        ).fetchall()
        block_texts = [b["content"] for b in blocks]

        center = best_center(
            event["name"], event["country"], event["description"], block_texts
        )

        try:
            fetch_map(center, output_path)
            print(f"  saved {output_path.name}  (center: {center!r})")
        except Exception as e:
            print(f"  ERROR {slug}: {e}")

    conn.close()


if __name__ == "__main__":
    main()
