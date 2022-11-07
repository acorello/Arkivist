import itertools as itt
from pathlib import *
from dataclasses import dataclass
from typing import *
import PyPDF2 as pdf
import isbnlib
from isbnlib import RE_ISBN13
import pipe


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


def books(folder: Path) -> Iterable[Path]:
    "Traverse folder looking book files"
    pdfs = folder.rglob("*.pdf")
    epubs = folder.rglob("*.epub")
    return itt.chain(pdfs, epubs)


def _isbn_of_pdf(file: Path):
    pdfr = pdf.PdfReader(file)
    map = pipe.map
    isbns = (
        pdfr.pages[:5]
        | map(pdf.PageObject.extract_text)
        | map(RE_ISBN13.finditer)
        | pipe.traverse
        | map(Match.group)
        | map(isbnlib.canonical)
    )
    return next(isbns, None)


def book_info(file: Path) -> BookInfo | None:
    "Extract ISBN from file; resolve Title, Year, Author, Publisher"
    if isbn := _isbn_of_pdf(file):
        return BookInfo(isbn=isbn, title=None, publisher=None, year=0, ext=None)


# Candidate Book: a PDF, an ePub, …
BASE_FOLDER = Path(Path.home(), "Documents", "Books, Magazines, Articles")


def organize(base_path: Path) -> None:
    for book_file in books(base_path):
        try:
            bookinfo = book_info(book_file)
            new_path = base_path.joinpath(path_by_isbn(bookinfo))
            print(f"Moving {book_file} to {new_path}")
        except Exception as e:
            print("Error processing", book_file, e)
            continue
