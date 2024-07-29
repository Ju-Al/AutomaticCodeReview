} // namespace sorting

/**
 * Test function
 */
static void test() {
    
    // example 1: vector ofi nt
    const int size1 = 7;
    std::cout<<"\nTest 1- as std::vector<int>...";
    std::vector <int> arr1 = {23, 10, 20, 11, 12, 6, 7};
    sorting::pancakeSort(arr1, size1);
    assert(std::is_sorted(arr1.begin(), arr1.end()));
    std::cout<<"Passed\n";
    for (int i = 0; i < size1; i++) {
        std::cout<<arr1[i]<<" ,";
    }
    std::cout<<std::endl;
    
    // example 2: vector of double
    const int size2 = 8;
    std::cout<<"\nTest 2- as std::vector<double>...";
    std::vector <double> arr2 = {23.56, 10.62, 200.78, 111.484,3.9, 1.2, 61.77, 79.6};
    sorting::pancakeSort(arr2, size2);
    assert(std::is_sorted(arr2.begin(), arr2.end()));
    std::cout<<"Passed\n";
    for (int i = 0; i < size2; i++) {
        std::cout<<arr2[i]<<", ";
    }
    std::cout<<std::endl;
  
    // example 3:vector of float
      const int size3=7;
    std::cout<<"\nTest 3- as std::vector<float>...";
    std::vector <float> arr3 = {6.56, 12.62, 200.78, 768.484, 19.27, 68.87, 9.6};
    sorting::pancakeSort(arr3, size3);
    assert(std::is_sorted(arr3.begin(), arr3.end()));
    std::cout<<"Passed\n";
    for (int i = 0; i < size3; i++) {
        std::cout<<arr3[i]<<", ";
    }
    std::cout<<std::endl;
	  
    
}

/**
 * Our main function with example of sort method.
 */
int main() {
    test();
    return 0;
}
