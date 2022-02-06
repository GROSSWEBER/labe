"""
Experimental.
"""

import luigi

from labe.base import Zstd, shellout
from labe.tasks import OpenCitationsSingleFile, Task


class ExpRefcatDownload(Task):
    """
    Download refcat v2; over 21h.
    """
    url = luigi.Parameter(default="https://archive.org/download/refcat_2022-01-03/refcat-doi-table-2022-01-03.json.zst")

    def run(self):
        output = shellout("""
                          curl -sL --retry 3 --fail {url} | zstd -c -d -T0 -q | LC_ALL=C sort -S50% | zstd -c -T0 > {output}
                          """,
                          url=self.url)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst", digest=True), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsDOITable(Task):
    """
    DOI to DOI table, sorted.
    """

    def requires(self):
        return OpenCitationsSingleFile()

    def run(self):
        output = shellout(r"""
                          zstdcat -T0 {input} |
                          cut -d, -f2,3 |
                          sed -e 's@,@\t@' |
                          LC_ALL=C sort -S 50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")
