package example5.order;

public class SMSAuthorizer {
    private boolean authenticated;

    public SMSAuthorizer() {
        this.authenticated = false;
    }

    public void verify_code(String code) {
        System.out.println("Verifying code");
        this.authenticated = true;
    }

    public boolean is_authentificated() {
        return this.authenticated;
    }
}
