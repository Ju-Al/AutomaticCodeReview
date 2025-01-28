package example4.shapes;

public class Square implements Shape {
    private double length;

    public Square(double length) {
        this.length = length;
    }

    @Override
    public double area() {
        return length * length;
    }

    @Override
    public double volume() {
        throw new UnsupportedOperationException("The volume cannot be calculated");
    }
}