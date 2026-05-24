
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

def gen_patterns(count: int, seed: int) -> list[str]:
    """Generate list of unique patterns."""
    random.seed(seed)
    patterns = set()

    while len(patterns) < count:
        
        
        if random.random() < 0.4:
        
            p = random.choice(WORDS) + " " + random.choice(WORDS)
        else:
        
            p = random.choice(WORDS)
        patterns.add(p)

    return sorted(patterns)  


def main():
    parser = argparse.ArgumentParser(description="Generate patterns for searching")
    parser.add_argument("--count", type=int, default=100, help="number of patterns")
    parser.add_argument("--out", required=True, help="output path")
    parser.add_argument("--seed", type=int, default=42, help="random seed")
    args = parser.parse_args()

    patterns = gen_patterns(args.count, args.seed)

    with open(args.out, "w") as f:
        for p in patterns:
            f.write(p + "\n")

    print(f"saved {len(patterns)} patterns → {args.out}", file=sys.stderr)


if __name__ == "__main__":
    main()