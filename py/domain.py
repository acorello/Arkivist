import fileformat
from dataclasses import dataclass
from pathlib import PurePath
import isbnlib


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


@dataclass(frozen=True)
class ISBN:
    ean13: str

    def __init__(self, isbn_like: str):
        self.ean13 = isbnlib.ean13(isbn_like)


# TODO: can I make this a sort of "interface"? (rather than throwing an exception, I force the implementation to implement the declared methods.)
class Book:
    def find_isbn() -> ISBN:
        raise NotImplementedError("should be implemented by file-specific subclasses")
