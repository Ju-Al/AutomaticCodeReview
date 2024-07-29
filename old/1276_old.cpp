{ 
	int ans = 1;	    	/// Initialize the answer to be returned 

	a = a % c;          	/// Update a if it is more than or 
				/// equal to c 

	if (a == 0) 
	{
	    return 0;       	/// In case a is divisible by c; 
	}
	
	while (b > 0) 
	{ 
		/// If b is odd, multiply a with answer 
		if (b & 1) 
		{
			ans = (ans*a) % c; 
		}
		
		/// b must be even now 
		b = b>>1;       // b = b/2 
		a = (a*a) % c; 
	}
	
	return ans; 
}
}  // namespace math 

/**
* @brief Main function
* @returns 0 on exit
*/
int main() 
{ 	
	/// Give two numbers num1, num2 and modulo m
	int num1 = 2; 
	int num2 = 5; 
	int m = 13; 
	
	std::cout << "The value of "<<num1<<" raised to exponent "<<num2<< 
	" under modulo "<<m<<" is " << math::power(num1, num2, m); 
	/// std::cout << "The value of "<<num1<<" raised to exponent "<<num2<<" 
	/// " under modulo "<<m<<" is " << power(num1, num2, m); 

	return 0;
} 

