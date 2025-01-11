package example4.printer;

import java.io.File;

public abstract class Printer {
    public abstract void print(File document);

    public abstract void fax(File document);

    public abstract void scan(File document);
}
