import subprocess
import requests
import os
import sys
import logging
from pathlib import Path
from requests.exceptions import RequestException
from typing import List, Optional


# Set up logging
logging.basicConfig(
   level=logging.INFO,
   format='%(asctime)s - %(levelname)s - %(message)s'
)


class TestGenerator:
   def __init__(self):
       self.api_key = os.getenv('OPENAI_API_KEY')
       self.model = os.getenv('OPENAI_MODEL', 'o1-preview')
      
       try:
           self.max_tokens = int(os.getenv('OPENAI_MAX_TOKENS', '10000'))
       except ValueError:
           logging.error("Invalid value for OPENAI_MAX_TOKENS. Using default value: 2000")
           self.max_tokens = 2000


       if not self.api_key:
           raise ValueError("OPENAI_API_KEY environment variable is not set")


   def get_changed_files(self) -> List[str]:
       """Retrieve list of changed files passed as command-line arguments."""
       if len(sys.argv) <= 1:
           return []
       return [f.strip() for f in sys.argv[1].split() if f.strip()]


   def detect_language(self, file_name: str) -> str:
       """Detect programming language based on file extension."""
       extensions = {
           '.py': 'Python',
           '.js': 'JavaScript',
           '.ts': 'TypeScript',
           '.java': 'Java',
           '.cpp': 'C++',
           '.cs': 'C#',
           '.go': 'Go'
       }
       _, ext = os.path.splitext(file_name)
       return extensions.get(ext.lower(), 'Unknown')


   def get_test_framework(self, language: str) -> str:
       """Get the appropriate test framework based on language."""
       frameworks = {
           'Python': 'pytest',
           'JavaScript': 'jest',
           'TypeScript': 'jest',
           'Java': 'JUnit',
           'C++': 'Google Test',
           'C#': 'NUnit',
           'Go': 'testing'
       }
       return frameworks.get(language, 'unknown')
  
   def get_related_files(self, language: str, file_name: str) -> List[str]:
       """Identify related files based on import statements or includes."""
       related_files = []
      
       try:
           if language in ["Python", "JavaScript", "TypeScript"]:
               with open(file_name, 'r') as f:
                   for line in f:
                       if 'import ' in line or 'from ' in line or 'require(' in line:
                           parts = line.split()
                           for part in parts:
                               # Handle relative imports
                               if part.startswith('.') and not part.startswith('..'):
                                   path = part.replace(".", "")
                                   for ext in ('.py', '.js', '.ts'):
                                       potential_file = f"{path}{ext}"
                                       if Path(potential_file).exists():
                                           related_files.append(potential_file)
                                           break
                               elif '.' in part:
                                   path = part.replace(".", "/")
                                   for ext in ('.py', '.js', '.ts'):
                                       potential_file = f"{path}{ext}"
                                       if Path(potential_file).exists():
                                           related_files.append(potential_file)
                                           break
                               else:
                                   if part.endswith(('.py', '.js', '.ts')) and Path(part).exists():
                                       related_files.append(part)
                                   elif part.isidentifier():
                                       base_name = part.lower()
                                       for ext in ('.py', '.js', '.ts'):
                                           potential_file = f"{base_name}{ext}"
                                           if Path(potential_file).exists():
                                               related_files.append(potential_file)
                                               break
           elif language == 'C++':
               # Implement C++ related file detection logic here
               pass  # Placeholder for C++ implementation
           elif language == 'C#':
               # Implement C# related file detection logic here
               pass  # Placeholder for C# implementation


       except Exception as e:
           logging.error(f"Error identifying related files in {file_name}: {e}")
      
       return related_files


   def get_related_test_files(self, language: str, file_name: str) -> List[str]:
       """Identify related test files based on import statements or includes."""
       related_test_files = []
       try:
           if language == "Python":
               directory = Path(os.path.dirname(os.path.abspath(__file__)))
               test_files = list(directory.rglob("tests.py")) + \
                            list(directory.rglob("test.py")) + \
                            list(directory.rglob("test_*.py")) + \
                            list(directory.rglob("*_test.py"))
               for test_file in test_files:
                   with open(test_file, 'r') as f:
                       for line in f:
                           if 'from ' in line:
                               parts = line.split()
                               for part in parts:
                                   if part.startswith('.') and not part.startswith('..'):
                                       path = part.replace(".", "")
                                       for ext in ('.py', '.js', '.ts'):
                                           potential_file = f"{path}{ext}"
                                           if Path(potential_file).exists() and potential_file in file_name:
                                               related_test_files.append(str(test_file))
                                               break
                                   elif '.' in part:
                                       path = part.replace(".", "/")
                                       for ext in ('.py', '.js', '.ts'):
                                           potential_file = f"{path}{ext}"
                                           if Path(potential_file).exists() and potential_file in file_name:
                                               related_test_files.append(str(test_file))
                                               break
                                   else:
                                       if part.endswith(('.py', '.js', '.ts')) and Path(part).exists() and (file_name in part):
                                           related_test_files.append(str(test_file))
                                       elif part.isidentifier():
                                           base_name = part.lower()
                                           for ext in ('.py', '.js', '.ts', '.js'):
                                               potential_file = f"{base_name}{ext}"
                                               if Path(potential_file).exists() and (file_name in potential_file):
                                                   related_test_files.append(str(test_file))
                                                   break
           # Implement related test file detection for other languages if needed


       except Exception as e:
           logging.error(f"Error identifying related test files in {file_name}: {e}")
      
       # Limit to the first related test file to avoid excessive processing
       limited_test_files = related_test_files[:1]
       return limited_test_files


   def generate_coverage_report(self, test_file: Path, language: str):
       """Generate a code coverage report and save it as a text file."""
       report_file = test_file.parent / f"{test_file.stem}_coverage_report.txt"


       try:
           if language == "Python":
               subprocess.run(
                   ["coverage", "run", str(test_file)],
                   check=True
               )
               subprocess.run(
                   ["coverage", "report", "-m", "--omit=*/site-packages/*"],
                   stdout=open(report_file, "w"),
                   check=True
               )
           elif language == "JavaScript":
               subprocess.run(
                   ["jest", "--coverage", "--config=path/to/jest.config.js"],
                   stdout=open(report_file, "w"),
                   check=True
               )
           elif language == "C++":
               # Implement coverage report generation for C++ using Google Test and gcov
               pass  # Placeholder for C++ coverage
           elif language == "C#":
               # Implement coverage report generation for C# using NUnit and coverlet
               pass  # Placeholder for C# coverage
           elif language == "Go":
               subprocess.run(
                   ["go", "test", "-coverprofile=coverage.out", str(test_file)],
                   check=True
               )
               subprocess.run(
                   ["go", "tool", "cover", "-html=coverage.out", "-o", str(report_file)],
                   check=True
               )
           else:
               logging.warning(f"No coverage report generation implemented for {language}.")


           logging.info(f"Code coverage report saved to {report_file}")


       except subprocess.CalledProcessError as e:
           logging.error(f"Error generating coverage report for {test_file}: {e}")


   def ensure_coverage_installed(self, language: str):
       """
       Ensures that the appropriate coverage tool for the given programming language is installed.
       Logs messages for each step.
       """
       try:
           if language.lower() == 'python':
               subprocess.check_call([sys.executable, '-m', 'pip', 'show', 'coverage'])
               logging.info("Coverage tool for Python is already installed.")
           elif language.lower() == 'javascript':
               subprocess.check_call(['npm', 'list', 'jest'])
               logging.info("Coverage tool for JavaScript (jest) is already installed.")
           elif language.lower() == 'java':
               logging.info("Ensure Jacoco is configured in your Maven/Gradle build.")
           elif language.lower() == 'ruby':
               subprocess.check_call(['gem', 'list', 'simplecov'])
               logging.info("Coverage tool for Ruby (simplecov) is already installed.")
           elif language.lower() == 'go':
               logging.info("Go coverage is handled by the 'go test' command.")
           else:
               logging.warning(f"Coverage tool check is not configured for {language}. Please add it manually.")
               return


       except subprocess.CalledProcessError:
           logging.error(f"Coverage tool for {language} is not installed. Installing...")
           try:
               if language.lower() == 'python':
                   subprocess.check_call([sys.executable, '-m', 'pip', 'install', 'coverage'])
                   logging.info("Coverage tool for Python has been installed.")
               elif language.lower() == 'javascript':
                   subprocess.check_call(['npm', 'install', 'jest'])
                   logging.info("Coverage tool for JavaScript (jest) has been installed.")
               elif language.lower() == 'ruby':
                   subprocess.check_call(['gem', 'install', 'simplecov'])
                   logging.info("Coverage tool for Ruby (simplecov) has been installed.")
               else:
                   logging.error(f"Could not install coverage tool for {language} automatically. Please install manually.")
           except subprocess.CalledProcessError:
               logging.error(f"Failed to install the coverage tool for {language}. Please install it manually.")


   def create_prompt(self, file_name: str, language: str) -> Optional[str]:
       """Create a language-specific prompt for test generation with accurate module and import names in related content."""
       try:
           with open(file_name, 'r') as f:
               code_content = f.read()
       except Exception as e:
           logging.error(f"Error reading file {file_name}: {e}")
           return None


       related_files = self.get_related_files(language, file_name)
       related_content = ""


       if related_files:
           logging.info(f"Related files for {file_name}: {related_files}")
       else:
           logging.info(f"No related files found for {file_name} to reference")


       for related_file in related_files:
           try:
               with open(related_file, 'r') as rf:
                   file_content = rf.read()
                   module_path = str(Path(related_file).with_suffix('')).replace('/', '.')
                   import_statement = f"import {module_path}"
                   related_content += f"\n\n// Module: {module_path}\n{import_statement}\n{file_content}"
                   logging.info(f"Included content from related file: {related_file} as module {module_path}")
           except Exception as e:
               logging.error(f"Error reading related file {related_file}: {e}")


       related_test_files = self.get_related_test_files(language, file_name)
       related_test_content = ""


       if related_test_files:
           logging.info(f"Related Test files for {file_name}: {related_test_files}")
       else:
           logging.info(f"No related test files found for {file_name} to reference")


       for related_test_file in related_test_files:
           try:
               with open(related_test_file, 'r') as rf:
                   file_content = rf.read()
                   related_test_content += f"\n\n// Related test file: {related_test_file}\n{file_content}"
                   logging.info(f"Included content from related test file: {related_test_file}")
           except Exception as e:
               logging.error(f"Error reading related test file {related_test_file}: {e}")


       framework = self.get_test_framework(language)
       prompt = f"""Generate comprehensive unit tests for the following {language} file: {file_name} using {framework}.


Requirements:
1. Include edge cases, normal cases, and error cases.
2. Use mocking where appropriate for external dependencies.
3. Include setup and teardown if needed.
4. Add descriptive test names and docstrings.
5. Follow {framework} best practices.
6. Ensure high code coverage.
7. Test both success and failure scenarios.


Code to test (File: {file_name}):


{code_content}


Related context:


{related_content}


Related test cases:


{related_test_content}


Generate only the test code without any explanations or notes."""


       logging.info(f"Created prompt for {file_name} with length {len(prompt)} characters")
       return prompt


   def call_openai_api_gptseries(self, prompt: str) -> Optional[str]:
     """Call OpenAI API to generate test cases."""
     headers = {
         'Content-Type': 'application/json',
         'Authorization': f'Bearer {self.api_key}'
     }
     
     data = {
         'model': self.model,
         'messages': [
             {
                 "role": "system",
                 "content": "You are a senior software engineer specialized in writing comprehensive test suites."
             },
             {
                 "role": "user",
                 "content": prompt
             }
         ],
         'max_tokens': self.max_tokens,
         'temperature': 0.7
     }

     try:
         response = requests.post(
             'https://api.openai.com/v1/chat/completions',
             headers=headers,
             json=data,
             timeout=60
         )
         response.raise_for_status()
         generated_text = response.json()['choices'][0]['message']['content']
         normalized_text = generated_text.replace('“', '"').replace('”', '"').replace("‘", "'").replace("’", "'")
         if normalized_text.startswith('```'):
             first_newline_index = normalized_text.find('\n', 3)
             if first_newline_index != -1:
                 normalized_text = normalized_text[first_newline_index+1:]
             else:
                 normalized_text = normalized_text[3:]
             if normalized_text.endswith('```'):
                 normalized_text = normalized_text[:-3]
         return normalized_text.strip()
     except RequestException as e:
         logging.error(f"API request failed: {e}, Response: {response.text}")
         return None
     
 
   def call_openai_api_mini(self, prompt: str) -> Optional[str]:
     """Call OpenAI API to generate test cases."""
     headers = {
         'Content-Type': 'application/json',
         'Authorization': f'Bearer {self.api_key}'
     }
     
     data = {
         'model': self.model,
         'messages': [
             {
                 "role": "user",
                 "content": prompt
             }
         ],
         'max_completion_tokens': self.max_tokens
     }

     try:
         response = requests.post(
             'https://api.openai.com/v1/chat/completions',
             headers=headers,
             json=data,
             timeout=60
         )
         response.raise_for_status()
         generated_text = response.json()['choices'][0]['message']['content']
         normalized_text = generated_text.replace('“', '"').replace('”', '"').replace("‘", "'").replace("’", "'")
         if normalized_text.startswith('```'):
             first_newline_index = normalized_text.find('\n', 3)
             if first_newline_index != -1:
                 normalized_text = normalized_text[first_newline_index+1:]
             else:
                 normalized_text = normalized_text[3:]
             if normalized_text.endswith('```'):
                 normalized_text = normalized_text[:-3]
         return normalized_text.strip()
     except RequestException as e:
         logging.error(f"API request failed: {e}, Response: {response.text}")
         return None
      
   def save_test_cases(self, file_name: str, test_cases: str, language: str) -> Optional[Path]:
       """Save generated test cases to appropriate directory structure."""
       tests_dir = Path('generated_tests')
       tests_dir.mkdir(exist_ok=True)
       lang_dir = tests_dir / language.lower()
       lang_dir.mkdir(exist_ok=True)
       base_name = Path(file_name).stem
       extension = Path(file_name).suffix


       # Adjust naming convention based on language
       if language.lower() == 'javascript':
           extension = '.js'
           if not base_name.startswith("test_"):
               base_name = f"test_{base_name}"
       elif language.lower() == 'go':
           extension = '.go'
           if not base_name.endswith("_test"):
               base_name = f"{base_name}_test"
           # Optionally, add any Go-specific headers or setup code if necessary
       elif language.lower() == 'python':
           if not base_name.startswith("test_"):
               base_name = f"test_{base_name}"


       test_file = lang_dir / f"{base_name}{extension}"
       header = ""


       if language.lower() == 'python':
           header = (
               "import sys\n"
               "import os\n"
               "sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../..')))\n\n"
           )
       elif language.lower() == 'go':
           # Include any Go-specific setup code here, if necessary
           test_dir = Path(file_name).parent
           test_file_name = f"{Path(file_name).stem}_test.go"


       try:
           with open(test_file, 'w', encoding='utf-8') as f:
               f.write(header + test_cases)
           logging.info(f"Test cases saved to {test_file}")
       except Exception as e:
           logging.error(f"Error saving test cases to {test_file}: {e}")
           return None


       if test_file.exists():
           logging.info(f"File {test_file} exists with size {test_file.stat().st_size} bytes.")
           return test_file
       else:
           logging.error(f"File {test_file} was not created.")
           return None


   def run(self):
       """Main execution method."""
       changed_files = self.get_changed_files()
       if not changed_files:
           logging.info("No files changed.")
           return


       for file_name in changed_files:
           if file_name == "generate_tests.py":
               continue  # Skip the test generation script itself
           try:
               language = self.detect_language(file_name)
               if language == 'Unknown':
                   logging.warning(f"Unsupported file type: {file_name}")
                   continue


               logging.info(f"Processing {file_name} ({language})")
               prompt = self.create_prompt(file_name, language)
              
               if prompt:
                   test_cases = self.all_openai_api_mini(prompt)
                  
                   if test_cases:
                       test_cases = test_cases.replace("“", '"').replace("”", '"')
                       self.ensure_coverage_installed(language)


                       test_file = self.save_test_cases(file_name, test_cases, language)
                       if test_file:
                           self.generate_coverage_report(test_file, language)
                   else:
                       logging.error(f"Failed to generate test cases for {file_name}")
           except Exception as e:
               logging.error(f"Error processing {file_name}: {e}")


if __name__ == '__main__':
   try:
       generator = TestGenerator()
       generator.run()
   except Exception as e:
       logging.error(f"Fatal error: {e}")
       sys.exit(1)
