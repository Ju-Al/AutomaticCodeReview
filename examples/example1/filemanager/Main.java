package example1.filemanager;

import java.io.IOException;

public class Main {
    public static void main(String[] args) {
        try {
            FileManager fileManager = new FileManager("filemanager/beispiel.txt");
            fileManager.write("Hello, this is a test.", "UTF-8");
            System.out.println("Inhalt der Datei: " + fileManager.read("UTF-8"));
            fileManager.compress();
            fileManager.decompress();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
