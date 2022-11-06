import itertools
from pathlib import *
from dataclasses import dataclass
import typing as ty


@dataclass(
    frozen=True,
)
class BookInfo:
    isbn: str
    title: str
    year: int
    publisher: str
    ext: str


def path_by_isbn(b: BookInfo) -> PurePath:
    return PurePath(
        "By-ISBN",
        f"{b.isbn} • {b.title} • {b.year} • {b.publisher}",
        f"{b.title}.{b.ext}",
    )


def path_by_publisher(b: BookInfo) -> PurePath:
    return PurePath(
        "By-Publisher",
        f"{b.isbn} • {b.title} • {b.year}",
        f"{b.title}.{b.ext}",
    )


def books(folder: Path) -> ty.Iterable[Path]:
    "Traverse folder looking book files"
    pdfs = folder.rglob("*.pdf")
    epubs = folder.rglob("*.epub")
    return itertools.chain(pdfs, epubs)


def resolve_metadata(file: Path) -> BookInfo:
    "Extract ISBN from file; resolve Title, Year, Author, Publisher"
    ...


# Candidate Book: a PDF, an ePub, …
BASE_FOLDER = Path(Path.home(), "Documents", "Books, Magazines, Articles")


def organize(base_path: Path) -> None:
    for book_file in books(base_path):
        try:
            bookinfo = resolve_metadata(book_file)
            new_path = base_path.joinpath(path_by_isbn(bookinfo))
            print(f"Moving {book_file} to {new_path}")
        except Exception as e:
            print("Error processing", book_file, e)
            continue
