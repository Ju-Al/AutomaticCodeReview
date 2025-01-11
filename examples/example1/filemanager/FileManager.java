package example1.filemanager;

import java.io.*;
import java.nio.file.*;
import java.util.zip.*;

public class FileManager {
    private Path path;

    public FileManager(String filename) {
        this.path = Paths.get(filename);
    }

    public String read(String encoding) throws IOException {
        return new String(Files.readAllBytes(path), encoding);
    }

    public void write(String data, String encoding) throws IOException {
        Files.write(path, data.getBytes(encoding));
    }

    public void compress() throws IOException {
        try (ZipOutputStream zipOut = new ZipOutputStream(
                new FileOutputStream(path.getFileName().toString() + ".zip"))) {
            File fileToZip = path.toFile();
            try (FileInputStream fis = new FileInputStream(fileToZip)) {
                ZipEntry zipEntry = new ZipEntry(fileToZip.getName());
                zipOut.putNextEntry(zipEntry);
                byte[] bytes = new byte[1024];
                int length;
                while ((length = fis.read(bytes)) >= 0) {
                    zipOut.write(bytes, 0, length);
                }
            }
        }
    }

    public void decompress() throws IOException {
        try (ZipInputStream zipIn = new ZipInputStream(new FileInputStream(path.getFileName().toString() + ".zip"))) {
            ZipEntry entry;
            while ((entry = zipIn.getNextEntry()) != null) {
                File outputFile = new File(entry.getName());
                try (BufferedOutputStream bos = new BufferedOutputStream(new FileOutputStream(outputFile))) {
                    byte[] bytesIn = new byte[1024];
                    int read;
                    while ((read = zipIn.read(bytesIn)) != -1) {
                        bos.write(bytesIn, 0, read);
                    }
                }
                zipIn.closeEntry();
            }
        }
    }
}
