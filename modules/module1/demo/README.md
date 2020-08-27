# Technology Review

## Anaconda

### Install Anaconda

Download from https://www.anaconda.com/products/individual

For Linux:

    apt update
    apt install -y libgl1-mesa-glx libegl1-mesa libxrandr2 libxrandr2 libxss1 libxcursor1 libxcomposite1 libasound2 libxi6 libxtst6 wget nano

    wget https://repo.anaconda.com/archive/Anaconda3-2020.07-Linux-x86_64.sh

    chmod +x Anaconda3-2020.07-Linux-x86_64.sh 

    ./Anaconda3-2020.07-Linux-x86_64.sh

    . ~/.bashrc

Other operating systems have graphical installers. If you are not using linux, run the "Anaconda Prompt" program after installing to get a command line.

### Create new Anaconda environment

https://docs.conda.io/projects/conda/en/latest/user-guide/tasks/manage-environments.html

    conda create --name ds498
    conda activate ds498

### Use Anaconda to install packages

List of available packages: https://anaconda.org/anaconda/repo

    conda install -c anaconda git jupyterlab


### Run jupyter lab

    jupyter lab

## Git exercise

* You have already installed git using the commands above. Keep in mind that it will only work when you are inside the conda environment `ds498`.
* Create a user account at https://github.com/.
* Read through the syllabus and find a problem with it.
* Go to https://github.com/uiceds/syllabus, click on "issues", and create a new issue that describes the problem. Make a note of the number assigned to the issue you just created.
* Now, fix the problem!
    * Go back to https://github.com/uiceds/syllabus and click the "fork" button at the top right to create your own personal github copy of the syllabus. 
    * In your copy, click the green "code" box to get the address to clone the syllabus repository.
    * Back in the command prompt, run the command `git clone 'address'` (with 'address' replaced with the actual address) to dowload the repository to your computer.
    * Navigate into the syllabus repository directory (for example, `cd syllabus`).
    * Create a new git branch where you will fix the problem: `git checkout -b bugfix`.
    * Fix the problem using a text editor. There are many text editors out there, but one good one is VSCode: https://code.visualstudio.com/.
    * Stage the change that you made using `git add path/to/file`, where 'path/to/file' is replaced by the path to the file that you changed.
    * Commit the change using `git commit -m "message"`, where the message describes the change you made.
    * Push your change back up to your github repository using `git push origin bugfix`.
    * Back in the web browser, go to your person version of the syllabus (https://github.com/yourname/syllabus), and create a pull request against the main version of the syllabus based on the change you just made. If you pushed your changes recently, there should be a banner at the top of the page to click on that does this, otherwise navigate to the 'branches' page.
    * Enter a succinct and polite description of the change you made. Make sure you include the text "Fixes #X" in the pull request description (where X is the number assigned to the issue you created) so that GitHub and link the pull request to the issue it fixes.
* Anyone that finds an error in the syllabus, creates an issue to describe it, and creates a pull request that fixes it before the discussion responses for Module 1 are due (next Tuesday) is exempt from having to write those discussion responses.



