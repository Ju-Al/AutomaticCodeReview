package example4.printer;

import java.io.File;

public class ModernPrinter extends Printer {

    @Override
    public void print(File document) {
        System.out.println("Printing " + document + " in black and white...");
    }

    @Override
    public void fax(File document) {
        System.out.println("Faxing " + document + "...");
    }

    @Override
    public void scan(File document) {
        System.out.println("Scanning " + document + "...");
    }

}
