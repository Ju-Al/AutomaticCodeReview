package example4.printer;

import java.io.File;

public class OldPrinter extends Printer {

    @Override
    public void print(File document) {
        System.out.println("Printing " + document + " in black and white...");
    }

    @Override
    public void fax(File document) {
        throw new UnsupportedOperationException("Not implemented");
    }

    @Override
    public void scan(File document) {
        throw new UnsupportedOperationException("Unimplemented method 'scan'");
    }

}
