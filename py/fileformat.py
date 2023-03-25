from pathlib import PurePath
from zipfile import ZipFile

import isbnlib
import isbnlib.registry
import magic
import PyPDF2 as pdf
from bs4 import BeautifulSoup
from domain import ISBN, Book


# Wrapper to extract info from book-files


class UnsupportedFormatException:
    def __init__(self, mime_type: str):
        self.message = mime_type


class Epub(Book):
    MIME_TYPE = "application/epub+zip"

    def __init__(self, filepath: PurePath | str):
        if filepath is str:
            filepath = PurePath(filepath)
        self.zip = ZipFile(filepath, mode="r")

    import re

    isbn_pattern = isbnlib.RE_STRICT.pattern[
        # remove leading ^ and trailing $ (RegEx anchors)
        1:-1
    ]
    isbn_prefix = re.compile(
        r"\s*(?:urn:)?(?:isbn[:_]?)?(" + isbn_pattern + r")\s*",
        re.IGNORECASE + re.UNICODE,
    )

    class XMLItem:
        def __init__(self, epub: ZipFile, itempath: str):
            self.epub = epub
            self.itempath = itempath

        @staticmethod
        def _make_soup(xml: str):
            return BeautifulSoup(xml, "xml")

        def _read_str(self):
            return str(self.epub.read(self.itempath), "utf-8")

        def xml(self):
            return self._read_str()

        def soup(self):
            return self._make_soup(self.xml())

    def _xml_item(self, itempath: str):
        return self.XMLItem(self.zip, itempath)

    def container(self):
        return self._xml_item("META-INF/container.xml")

    def rootfile(self):
        rootfile_path = self.container().soup().rootfile["full-path"]
        return self._xml_item(rootfile_path)

    def identifier(self):
        return self.rootfile().soup().identifier.text

    def xml_item(self, item_id: str):
        item_path = (
            self.rootfile()
            .soup()
            .find("item", **{"id": item_id, "media-type": "application/xhtml+xml"})[
                "href"
            ]
        )
        if item_path:
            return self._xml_item(item_path)

    def xml_items(self):
        return (
            self.rootfile()
            .soup()
            .find_all("item", **{"media-type": "application/xhtml+xml"})
        )

    def find_isbn(self):
        if m := self.isbn_prefix.match(self.identifier()):
            return m[0]
        else:
            return None


class Pdf(Book):
    MIME_TYPE = "application/pdf"

    def __init__(self, path: PurePath):
        self.pdf_reader = pdf.PdfReader(path)

    def find_isbn(self):
        "The first ISBN-like string found within the first 5 pages or None."
        pages = self.pdf_reader.pages[:5]
        text = "".join([pdf.PageObject.extract_text(p) for p in pages])
        isbn: str = next(isbnlib.get_isbnlike(text))
        return ISBN.of(isbn) if isbn else None


def mime_type(filepath: PurePath) -> str | None:
    with magic.Magic(flags=magic.MAGIC_MIME_TYPE) as m:
        return m.id_filename(filepath.as_posix())


def load_book(file: PurePath) -> Book | None:
    match mime_type(file):
        case Pdf.MIME_TYPE:
            return Pdf(file)
        case Epub.MIME_TYPE:
            return Epub(file)
        case _:
            raise UnsupportedFormatException(mime_type)
