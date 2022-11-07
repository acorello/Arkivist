import itertools as itt
from pathlib import *
from dataclasses import dataclass
from typing import *
import PyPDF2 as pdf
import isbnlib
from isbnlib.registry import PROVIDERS as METADATA_PROVIDERS
from isbnlib import ISBNLibException
from pipe import traverse, map as fn
import magic
import mimetypes


class ArkivistException(Exception):
    def __str__(self):
        return getattr(self, "message", "")  # pragma: no cover


class MetadataRetrievalFailed(ArkivistException):
    def __init__(self, message):
        self.message = message


class UnsupportedFormat(ArkivistException):
    def __init__(self, format_str: str):
        self.message = f"Format: {format_str}"


def _isbn_of_pdf(file: Path):
    pdfr = pdf.PdfReader(file)
    isbns = (
        pdfr.pages[:5]
        | fn(pdf.PageObject.extract_text)
        | fn(isbnlib.get_isbnlike)
        | traverse
        | fn(isbnlib.canonical)
    )
    return next(isbns, None)


def _book_metadata(isbn: str) -> dict[str, str]:
    data = None
    for provider in METADATA_PROVIDERS:
        try:
            if data := isbnlib.meta(isbn, service=provider):
                return data
        except ISBNLibException:
            continue
    raise MetadataRetrievalFailed(f"ISBN: {isbn}")


def _file_extension(filepath: PurePath) -> str | None:
    with magic.Magic(flags=magic.MAGIC_MIME_TYPE) as m:
        mimetype = m.id_filename(filepath.as_posix())
        if ext := mimetypes.guess_extension(mimetype):
            return ext[1:]  # strip leading dot


def _isbn(file: PurePath, file_format: str):
    match file_format:
        case "pdf":
            return _isbn_of_pdf(file)
        case _:
            raise UnsupportedFormat(file_format)


@dataclass(
    frozen=True,
)
class BookInfo:
    isbn: str
    title: str
    year: int
    publisher: str
    extension: str


def path_by_isbn(b: BookInfo) -> PurePath:
    return PurePath(
        "By-ISBN",
        f"{b.isbn} • {b.title} • {b.year} • {b.publisher}",
        f"{b.title}.{b.extension}",
    )


def path_by_publisher(b: BookInfo) -> PurePath:
    return PurePath(
        "By-Publisher",
        f"{b.isbn} • {b.title} • {b.year}",
        f"{b.title}.{b.extension}",
    )


def books(folder: Path) -> Iterable[Path]:
    "Traverse folder looking book files"
    return ["*.pdf", "*.epub"] | fn(folder.rglob) | traverse


def book_info(file: Path) -> BookInfo | None:
    "Extract ISBN from file; resolve Title, Year, Author, Publisher"
    ext = _file_extension(file)
    if isbn := _isbn(file, ext):
        m = _book_metadata(isbn)
        return BookInfo(
            isbn=m["ISBN-13"],
            title=m["Title"].title(),
            publisher=m["Publisher"].title(),
            year=int(m["Year"]),
            extension=ext,
        )


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
