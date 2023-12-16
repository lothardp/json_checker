#!/bin/zsh

# compile
go build main.go

failed_tests=()

echo "Running coding challenge tests..."
for file in ./tests/**/*.json; do
	cat $file | ./main > /dev/null 2>&1

	if [ $? -eq 0 ]; then
		if [[ $file == *"invalid"* ]]; then
			SYMBOL="✗"
			failed_tests+=($file)
		else
			SYMBOL="✓"
		fi
		echo "$SYMBOL Test passed: $file"
	else
		if [[ $file == *"invalid"* ]]; then
			SYMBOL="✓"
		else
			SYMBOL="✗"
			failed_tests+=($file)
		fi
		echo "$SYMBOL Test failed: $file"
	fi
done

echo "Running official tests..."
for file in ./test/*.json; do
	cat $file | ./main > /dev/null 2>&1

	if [ $? -eq 0 ]; then
		if [[ $file == *"fail"* ]]; then
			SYMBOL="✗"
			failed_tests+=($file)
		else
			SYMBOL="✓"
		fi
		echo "$SYMBOL Test passed: $file"
	else
		if [[ $file == *"fail"* ]]; then
			SYMBOL="✓"
		else
			SYMBOL="✗"
			failed_tests+=($file)
		fi
		echo "$SYMBOL Test failed: $file"
	fi
done

echo ""

if [ ${#failed_tests[@]} -eq 0 ]; then
	echo "All tests passed!"
else
	echo "Failed tests:"
	for test in $failed_tests; do
		echo $test
	done
fi
