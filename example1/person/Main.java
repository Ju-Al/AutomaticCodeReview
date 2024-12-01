package example1.person;

public class Main {
    public static void main(String[] args) {
        ValidatePerson person = new ValidatePerson("John", 20);
        person.display();

        ValidatePerson invalidPerson = new ValidatePerson("Jo", 17);
        invalidPerson.display();
    }
}
