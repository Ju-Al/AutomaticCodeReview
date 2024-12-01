package example3.geometry.java;

public class Line {
    private Point p1;
    private Point p2;

    public Line(Point p1, Point p2) {
        this.p1 = p1;
        this.p2 = p2;
    }

    public double getSlope() {
        if (p2.getX() == p1.getX()) {
            throw new ArithmeticException("Slope is undefined for a vertical line.");
        }
        return (p2.getY() - p1.getY()) / (p2.getX() - p1.getX());
    }

    public double getIntercept() {
        return p1.getY() - (getSlope() * p1.getX());
    }

    public Point getP1() {
        return p1;
    }

    public Point getP2() {
        return p2;
    }

    public boolean isOn(Point point) {
        // Checks if the point lies on the infinite line
        double slope = getSlope();
        return Double.compare(point.getY(), slope * point.getX() + getIntercept()) == 0;
    }
}
