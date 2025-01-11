package example4.shapes;

public class Dice implements Shape {
    private double length;

    public Dice(double length) {
        this.length = length;
    }

    @Override
    public double area() {
        return 6 * Math.pow(length, 2);
    }

    @Override
    public double volume() {
        return Math.pow(length, 3);
    }
}
