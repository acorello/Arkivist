from dataclasses import dataclass
from pathlib import Path, PurePath
from typing import Iterable

import isbnlib
import isbnlib.registry

from . import fileformat
from . import metadata_source

# I like the java naming convention: fully qualified names make it obvious to see which imports are from third parties.


@dataclass
class NoISBN:
    pass


@dataclass(frozen=True)
class BookInfo:
    isbn: str
    title: str
    year: int
    publisher: str


@dataclass(frozen=True)
class BookFileInfo(BookInfo):
    mime_type: str

    def path_by_isbn(self) -> PurePath:
        return PurePath(
            "By-ISBN",
            f"{self.isbn} • {self.title} • {self.year} • {self.publisher}",
            f"{self.title}.{self.extension()}",
        )

    def extension(self):
        match self.mime_type:
            case fileformat.Epub.MIME_TYPE:
                "epub"
            case fileformat.Pdf.MIME_TYPE:
                "pdf"


class ISBN:
    def __init__(self, isbn_like: str):
        self.ean13 = isbnlib.ean13(isbn_like)

    def __str__(self):
        return self.ean13

    def canonical(self):
        return isbnlib.canonical(self.ean13)


# TODO: can I make this a sort of "interface"? (rather than throwing an exception, I force the implementation to implement the declared methods.)
class Book:
    def find_isbn() -> ISBN:
        raise NotImplementedError("should be implemented by file-specific subclasses")


def current_book_paths(folder: Path) -> Iterable[Path]:
    """Paths to each book in folder (*.{pdf,epub} files).

    In case of multiple links to the same file only the first is returned."""
    for exts in ["*.pdf", "*.epub"]:
        for path in folder.rglob(exts):
            if path.is_file():
                yield path


def book_info(file: Path):
    "Extract ISBN from file; resolve Title, Year, Author, Publisher"
    book = fileformat.load_book(file)
    if isbn := book.find_isbn():
        return metadata_source.get_book_info(isbn)
    else:
        return NoISBN()


def paths_from_bookinfo(book_info: BookInfo):
    # TODO: generate other paths, eg By-Publisher, By-Topic, etc.
    return [
        book_info.path_by_isbn(),
    ]


# file -> book_info   -> ProperPaths
#         no_isbn     -> NoISBNPath
#         no_metadata -> NoMetadataPath


def proper_book_paths(base_path: PurePath, book_path: PurePath) -> list[PurePath]:
    "Given a file Returns relative paths under which the book should be filed"

    def path_under(sub_folder: str):
        return sub_folder.joinpath(book_path.relative_to(base_path))

    match book_info(book_path):
        case BookInfo() as b:
            return paths_from_bookinfo(b)
        case NoISBN():
            return [path_under("No-ISBN")]
        case metadata_source.NoMetadata():
            return [path_under("No-Metadata")]


BASE_FOLDER = Path(Path.home(), "Documents", "Books and Articles")
TARGET_FOLDER = BASE_FOLDER.with_suffix(".ARKIVIST")


def organize_books(base_path: Path) -> None:
    assert base_path.exists() and base_path.is_dir() and base_path.is_absolute()
    for book_path in current_book_paths(base_path):
        # I'm running this on a folder having multiple paths pointing to the same file.
        # If books naively returns all paths that resolve to the same file, the same file will be
        #   processed multiple times; which is inefficient.
        # Symlinks or Hardlinks?
        # - hardlinks for files within the same file-system
        # - symlinks for files across file systems or for directories
        try:
            return book_path, list(
                map(lambda p: TARGET_FOLDER.joinpath(p), (base_path, book_path))
            )
        except Exception as e:
            print("Error processing", book_path, e)
            continue
