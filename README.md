# Run-Change
Run a command when a file is changed

*This is a Golang clone of Python's [when-changed](https://github.com/joh/when-changed)*




### What is it?

Tired of switching to the shell to test the changes you just made to
your code? Starting to feel like a mindless drone, manually running
pdflatex for the 30th time to see how your resume now looks?

Worry not, when-changed is here to help! Whenever it sees that you have
changed the file, when-changed runs any command you specify.

So to generate your latex resume automatically, you can do this:

```sh
$ rc CV.tex pdflatex CV.tex
```

Sweetness!




### Installation

```sh

```


### Usage

```sh
rc [OPTION] FILE COMMAND...
rc [OPTION] FILE [FILE ...] -c COMMAND
```

FILE can be a directory. Use %f to pass the filename to the command.

**Options:**

- -r Watch recursively
- -v Verbose output. Multiple -v options increase the verbosity. The maximum is 3: -vvv.
- -1 Don't re-run command if files changed while command was running
- -s Run command immediately at start
- -q Run command quietly

### Environment variables:

when-changed provides the following environment variables:

- WHEN_CHANGED_EVENT: reflects the current event type that occurs.
Could be either:
  - file_created
  - file_modified
  - file_moved
  - file_deleted

- WHEN_CHANGED_FILE: provides the full path of the file that has generated the event.
