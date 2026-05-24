
import argparse
import random
import sys


WORDS = [
    "the", "be", "to", "of", "and", "a", "in", "that", "have", "it",
    "for", "not", "on", "with", "he", "she", "as", "you", "do", "at",
    "this", "but", "his", "by", "from", "they", "we", "say", "her", "hers",
    "or", "an", "will", "my", "one", "all", "would", "there", "their",
    "what", "so", "up", "out", "if", "about", "who", "get", "which", "go",
    "when", "make", "can", "like", "time", "no", "just", "him", "know",
    "take", "people", "into", "year", "your", "good", "some", "could",
    "them", "see", "other", "than", "then", "now", "look", "only", "come",
    "its", "over", "think", "also", "back", "after", "use", "two", "how",
    "our", "work", "first", "well", "way", "even", "new", "want", "because",
]

def gen_text(target_bytes: int, seed: int) -> bytes:    
    random.seed(seed)
    chunks = []
    total = 0

    while total < target_bytes:
        line_len = random.randint(5, 15)
        words = [random.choice(WORDS) for _ in range(line_len)]
        line = " ".join(words) + "\n"
        encoded = line.encode("utf-8")
        chunks.append(encoded)
        total += len(encoded)

    return b"".join(chunks)


def parse_size(s: str) -> int:
    s = s.lower().strip()
    if s.endswith("gb"):
        return int(s[:-2]) * 1_000_000_000
    if s.endswith("mb"):
        return int(s[:-2]) * 1_000_000
    if s.endswith("kb"):
        return int(s[:-2]) * 1_000
    return int(s)  


def main():
    parser = argparse.ArgumentParser(description="Generate synthetic text file")
    parser.add_argument("--size", default="100mb", help="file size (e.g., 10mb, 1gb)")
    parser.add_argument("--out", required=True, help="output path")
    parser.add_argument("--seed", type=int, default=42, help="random seed")
    args = parser.parse_args()

    target = parse_size(args.size)
    print(f"generating {args.size} ({target:,} bytes) → {args.out}", file=sys.stderr)

    data = gen_text(target, args.seed)

    with open(args.out, "wb") as f:
        f.write(data)

    print(f"saved {len(data):,} bytes", file=sys.stderr)


if __name__ == "__main__":
    main()