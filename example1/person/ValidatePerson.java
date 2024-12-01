package example1.person;

public class ValidatePerson {

    private final String name;
    private final int age;

    public ValidatePerson(String name, int age) {
        this.name = name;
        this.age = age;
    }

    public boolean validateName(String name) {
        return name.length() > 3;
    }

    public boolean validateAge(int age) {
        return age > 18;
    }

    public void display() {
        if (validateName(this.name) && validateAge(this.age)) {
            System.out.println("Name: " + this.name + " and Age: " + this.age);
        } else {
            System.out.println("Invalid");
        }
    }
}