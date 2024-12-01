package example3.shapes.java;

public class Square implements Shape {
    private double length;

    public Square(double length) {
        this.length = length;
    }

    @Override
    public double area() {
        return length * length;
    }
}