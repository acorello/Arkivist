import os
import itertools as itt
from datetime import datetime
from pathlib import *
from dataclasses import dataclass
from typing import *
import PyPDF2 as pdf
import isbnlib
from isbnlib.registry import PROVIDERS as METADATA_PROVIDERS
from isbnlib import ISBNLibException
from pipe import traverse, map as fn
import pipe
import magic
import mimetypes

# What can go wrong?
# The path constructed based on the file metadata may be the same as the file itself
# so I could get an exception trying to write the file onto itself; if I go parallel
# or concurrent, I could have different files trying to write on each other.
# Depending on the sequential or concurrent execution model, I could adopt different
# strategies to prevent overriding the file.

# For the moment I'm going to implement the algorithm to:
# - move recognized files in the desired location
# - move unrecognized files in a special location


class ArkivistException(Exception):
    def __str__(self):
        return getattr(self, "message", "")  # pragma: no cover


class MetadataRetrievalException(ArkivistException):
    def __init__(self, message):
        self.message = message


class UnsupportedFormatException(ArkivistException):
    def __init__(self, format_str: str):
        self.message = f"Format: {format_str}"


@dataclass
class NoISBN:
    pass


@dataclass
class NoMetadata:
    pass


class ISBN:
    def __init__(self, isbn_like: str):
        self.__ean13 = isbnlib.ean13(isbn_like)

    def as_str(self):
        return self.__ean13

    def canonical(self):
        return isbnlib.canonical(self.__ean13)


def _isbn_of_pdf(file: Path) -> ISBN | None:
    "The first ISBN-like string found within the first 5 pages or None."
    return next(
        pdf.PdfReader(file).pages[:5]
        | pipe.map(pdf.PageObject.extract_text)
        | pipe.map(isbnlib.get_isbnlike)
        | traverse
        | pipe.map(ISBN),
        None,
    )


def _book_metadata(isbn: ISBN) -> dict[str, str]:
    # No, don't put this function within ISBN. We should do data-oriented functional programming.
    data = None
    # TODO: make this concurrent or parallel; does isbnlib already provides a method for that?
    for provider in METADATA_PROVIDERS:
        try:
            if data := isbnlib.meta(isbn.as_str(), service=provider):
                return data
        except ISBNLibException:
            continue
    raise MetadataRetrievalException(f"ISBN: {isbn}")


def book_format(filepath: PurePath) -> str | None:
    with magic.Magic(flags=magic.MAGIC_MIME_TYPE) as m:
        mimetype = m.id_filename(filepath.as_posix())
        if ext := mimetypes.guess_extension(mimetype):
            return ext[1:]  # strip leading dot


def _isbn(file: PurePath, file_format: str) -> ISBN | None:
    match file_format:
        case "pdf":
            return _isbn_of_pdf(file)
        case _:
            raise UnsupportedFormatException(file_format)


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


def candidate_books(folder: Path) -> Iterable[Path]:
    """Paths to each book in folder (*.{pdf,epub} files).

    In case of multiple links to the same file only the first is returned."""
    seen_inodes = set()
    for exts in ["*.pdf", "*.epub"]:
        for path in folder.rglob(exts):
            inode = os.stat(path).st_ino
            if not (inode in seen_inodes):
                seen_inodes.add(inode)
                yield path


def book_info(file: Path) -> BookInfo | NoISBN | NoMetadata:
    "Extract ISBN from file; resolve Title, Year, Author, Publisher"
    formt = book_format(file)
    # 1. book_format(path) -> PDF | ePub | … | None(it's not a book)
    #    1.a book_isbn(path,format) -> ISBN | None
    #    1.a book_metadata(isbn) -> BookMetadata | Error | None
    if isbn := _isbn(file, formt):
        if m := _book_metadata(isbn):
            return BookInfo(
                isbn=m["ISBN-13"],
                title=m["Title"].title(),
                publisher=m["Publisher"].title(),
                year=int(m["Year"]),
                extension=formt,
            )
        else:
            return NoMetadata()
    else:
        return NoISBN()


# Candidate Book: a PDF, an ePub, …
BASE_FOLDER = Path(Path.home(), "Documents", "Books and Articles")


def _identical_files(left: PurePath, right: PurePath) -> bool:
    stat = lambda f: os.stat(f, follow_symlinks=True)
    stat_l = stat(left)
    stat_r = stat(right)

    def same_content(left, right):
        if stat_l.st_size != stat_r.st_size:
            return False
        open = lambda f: Path.open(f, "rb")
        with open(left) as l, open(right) as r:
            buffer_size = 2**20  # 1MB
            while (l_buf := l.read(buffer_size)) or (r_buf := r.read(buffer_size)):
                if r_buf != l_buf:
                    return False
            return True

    return Path.samefile(left, right) or same_content(left, right)


# I have a folder (base_path) with a bunch of ebooks I want to organize in a specific folder structure.
#
# Approach: reorganize them within the base_path.
#   case 'a file is already in its place'.
#   case 'two files are classified as the same'.
#
#   Complexity 1.a: I have to strip 'Unresolved' from


def assert_metadata_preserved(before, after):
    assert before == after


def _fs_metadata(f: Path):
    # TODO: retrieve filesystem metadata
    None


NO_ISBN_FOLDER = PurePath("No-ISBN")  # must be relative
NO_METADATA_FOLDER = PurePath("No-Metadata")  # must be relative


def _add_suffix(some_path: PurePath) -> PurePath | None:
    for i in range(1, 10):
        alt_path = some_path.with_name(f"{some_path.stem}_{i}{some_path.suffix}")
        if not alt_path.exists():
            return alt_path
    return None


@dataclass
class MoveBookTo:
    new_path: PurePath


@dataclass
class BookAlreadyInPlace:
    pass


def organize_book(base_path: PurePath, book_path: PurePath) -> PurePath:
    def _subfolder_if_different(subfolder: PurePath):
        relative_book_path = book_path.relative_to(base_path)
        if relative_book_path.parts[0] == subfolder.as_posix():
            return None  # already in place
        else:
            return base_path.joinpath(NO_ISBN_FOLDER, relative_book_path)

    def _book_new_path() -> Path | None:
        match book_info(book_path):
            case BookInfo() as b:
                return b | fn(path_by_isbn) | fn(base_path.joinpath)
            case NoISBN():
                return _subfolder_if_different(NO_ISBN_FOLDER)
            case NoMetadata():
                return _subfolder_if_different(NO_METADATA_FOLDER)

    def _replace(book_new_path: PurePath):
        fs_metadata_before = _fs_metadata(book_path)
        book_path.replace(book_new_path)
        fs_metadata_after = _fs_metadata(book_new_path)
        assert_metadata_preserved(fs_metadata_before, fs_metadata_after)

    if book_new_path := _book_new_path():
        # book_new_path(s) generates the paths under which a book should be filed
        # one or more of such paths may be
        # - already used by the book itself, if it is already filed in the correct location
        # - already used by another book
        # - free

        # if the path is free it should be used.
        # if one of the paths corresponds to the book itself than
        #   we should not unlink the file.
        # else
        #   we should unlink the file (it was filed in an invalid location).
        # if one of the paths exists but is not identical to the book itself
        def equivalent_files(file1, file2):
            "file1 has the same information of file2 (from an end-user perspective); ie metadata may be different, encoding may be different"
            pass

        def identical_files(file1, file2) -> bool:
            "file1 is byte-by-byte the same as file2"
            pass

        def identical_paths(path1, path2) -> bool:
            "path1 is char-by-char the same as path2"
            pass

        def equivalent_paths(path1, path2) -> bool:
            "path1 refers to the same file referred by path2"
            pass

        # two files with the same ISBN should imply file-equivalence because:
        # - different ebook formats => different ISBN
        # - different book versions => different ISBN
        if book_new_path.exists() and not _identical_files(book_path, book_new_path):
            book_new_path = _add_suffix(book_new_path)
            if not book_new_path:
                return  # TODO: report too many copies with suffix
        else:  # same path, identical file
            return
    if book_new_path:
        _replace(book_new_path)


def organize_books(base_path: Path) -> None:
    assert base_path.exists() and base_path.is_dir() and base_path.is_absolute()
    for book_path in candidate_books(base_path):
        # I'm running this on a folder having multiple paths pointing to the same file.
        # If books naively returns all paths that resolve to the same file, the same file will be
        #   processed multiple times; which is inefficient.
        # Symlinks or Hardlinks?
        # - hardlinks for files within the same file-system
        # - symlinks for files across file systems or for directories
        try:
            organize_book(base_path, book_path)
        except Exception as e:
            print("Error processing", book_path, e)
            continue
