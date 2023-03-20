from pathlib import Path
import re
from typing import List, Iterable
import sys


RE_CA = re.compile("^N: (\d+)\n.+\n.+\n.+\ncounter(\d+)\n(\d+.\d+), (\d+.\d+), (\d+.\d+)\n.+\n.+\n.+\ncounter(\d+)\n(\d+.\d+), (\d+.\d+), (\d+.\d+)", re.MULTILINE)

RE_X = re.compile("^N: (\d+)\n.+\n.+\ncounter(\d+)\n(\d+.\d+), (\d+.\d+), (\d+.\d+)\n.+\n.+\ncounter(\d+)\n(\d+.\d+), (\d+.\d+), (\d+.\d+)", re.MULTILINE)


def format_csv(raw_iter: Iterable[tuple]) -> List[str]:
    lines: List[str] = []
    for raw in raw_iter:
        lines.append(f"{raw[0]}, {raw[1]}, {raw[2]}, {raw[3]}, {raw[4]}")
        lines.append(f"{raw[0]}, {raw[5]}, {raw[6]}, {raw[7]}, {raw[8]}")

    return lines


def parse_file(raw_emp: Path, is_ca: bool) -> List[str]:
    regex = RE_CA if is_ca else RE_X
    content = raw_emp.read_text()
    return format_csv(match.groups() for match in regex.finditer(content))


def main() -> None:
    lines = parse_file(Path(sys.argv[1]), "ca" in sys.argv[1])
    for line in lines:
        print(line)


if __name__ == "__main__":
    main()