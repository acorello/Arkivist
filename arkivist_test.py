import arkivist as arkv
from pathlib import Path, PurePath
import datetime as dt
import shutil as sh
import pytest as pyt
from pytest import fixture

"""markdown
# Interim Goal 1 (interim because selecting by extension is not ideal: the type should be implied by the file contents not by the file name)

Given a folder with electronics books stored as `*.pdf` or `*.epub` files in various sub-folders,

I want to file them in different categories, each category represented by a path.

I want to use the following categories:
- `$BASE/By-ISBN/{ISBN} • {Title} • {Year} • {Publisher}/{Title}.{ext}`
- `$BASE/No-ISBN/{Title} • {Authors} • {Year} • {Publisher}/{Title}.{ext}`
- `$BASE/By-Publisher/{Title} • {Year}/{Title}.{ext}`
- …a number of other categories 

Files in the same folder are different formats of that book.

If a book is saved only with one format, it will still be placed within a folder.

I wish I could create hard-links to link to a book from different categories, but Apple's readers don't work well with it.
So I'm falling back to symlinks to the book folder (or should I use symlinks to the book files?).

Folders are categories, sub-folders are sub-categories. Absolute paths are categorization.
A book is linked under different paths.
Possible LinkingStrategies:
- hard-links to files
- symlinks to files
- symlinks to folder
- OS-specific alias to file
- OS-specific alias to folder


Testing Strategy:
- create a sample-folder with:
  - {pdf,epub} book {with,without} ISBN in the {correct,incorrect} place
  # I'm reluctant to hardcode the enumerated set, it's a duplication; but how can I map each testcase with the expected result if I don't do that?
  # should I consider books for which I only have incomplete metadata?
    // target-place: By-ISBN|No-ISBN, By-Publisher, By-Author.
"""

# TODO: conflicting knowledge:
#       - each book format (PDF, ePub, paper) has a different ISBN
#       - multiple formats are stored within the same folder under By-ISBN category

NO_ISBN_FOLDER = PurePath("No-ISBN")


@fixture
def books_folder(pytestconfig: pyt.Config) -> Path:
    sample_books = pytestconfig.rootpath.joinpath("sample_books")
    samples_copy = Path().joinpath(
        "tmp", f"sample_books_{dt.datetime.now().isoformat()}"
    )
    sh.copytree(sample_books, samples_copy)
    yield samples_copy
    sh.rmtree(samples_copy)


def sanity_check_test(books_folder: Path):
    assert books_folder.exists()
    assert sum(1 for _ in arkv.candidate_books(books_folder)) == 4  # is this useful?


@fixture
def pdf_without_metadata(pytestconfig: pyt.Config):
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-ISBN",
        "978-1-098-10483-2 ÷ Knowledge Graphs - Data in Context for Responsive Businesses ÷ O'Reilly",
        "book.pdf",
    )


@pyt.mark.skip(reason="TODO")
def pdf_without_metadata_test(pdf_without_metadata: PurePath):
    ...


@fixture
def pdf_with_metadata(pytestconfig: pyt.Config):
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-Publisher",
        "O'Reilly",
        "Erlang Programming",
        "Erlang Programming - Francesco Cesarini.pdf",
    )


@fixture
def book_without_isbn():
    ...


@pyt.mark.skip(reason="TODO")
def books_without_isbn_test(book_without_isbn: PurePath):
    f"Books without ISBN are placed under {NO_ISBN_FOLDER}"
    ...


@pyt.mark.skip(reason="TODO")
def other_files_beside_books_test():
    ...


def resolve_metadata_test(pdf_with_metadata: Path):
    assert arkv.book_info(pdf_with_metadata) == arkv.BookInfo(
        isbn="9780596518189",
        title="Erlang Programming",
        year=2009,
        publisher="O'Reilly",
        extension="pdf",
    )
