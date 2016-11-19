# Tasks

A Task is a single action or command inside a script.

### Supported Tasks

### Future Tasks
- **get** Make an HTTP GET request
- **post** Make an HTTP POST request
- **extract** Extract data from the page and save as a variable
  - *regex* Extract text matching a regular expression
  - *element* Extract text matching a jquery selector
  - *json* Extract a JSON value at path
  - *as* Variable to save the text in. If not specified, it is saved in the
    Extracted array
- **sleep** Wait specified number of seconds
  - *seconds* How many seconds to sleep 
- *set* Set a variable during script execution.
  - *value* The value of the variable
  - *as* The name of the variable to set. Alias: *key*


### Variables
Gidra supports setting script-global variables that persist for the entire
execution of the script, as well as local variables that persist for a single
iteration of the loop, which goes through all of the tasks in the *tasks* list
once. Variables are accessible in the `Vars` map using the standard go template
syntax. 

Variables that are set inside the main `tasks` list persist only for a single
iteration of the loop. Variables set in the `vars`, `begin`, or `finally`
sections persist for all iterations of the loop.

### Conditions
All tasks can be made conditional with the *when* parameter. The task will only
be executed if the value of *when* evaluates to true. The value of when is a
standard Go template, allowing use of all previously defined variables and other
Go template logic.

### Outcomes
All tasks can be evaluated to see if they succeeded, determining whether the
loop should keep executing its tasks or continue to the next iteration of the
loop. Outcomes are evaluated similarly to a case statement, in the order *success-abort-retry-fail*. If one
condition is valid, the others are not checked, and execution advances. Syntax
for the *when* paramters supports all Go template logic. The opening `{{` and
closing `}}` brackets are not required in *when* and *with* parameters because
the code in them is always executed as a template.
- **success** The task has executed successfully, and execution should advance
  to the next task in the list. This is assumed as the default if no outcomes
  are specified.
    - *when* Condition that should evaluate to true to trigger this outcome.
- **abort** Something has gone wrong, and execution of the **entire script**
  should stop. Abort functions similarly to an unhandled panic or an assertion.
    - *when*
- **retry** The task should be retried again, optionally after executing the
  command in *with*. Retries will be attempted up to *limit* times. If *limit*
  is not set, it defaults to 5.
    - *with* Code to execute before retrying the task
    - *limit* How many times to retry before considering the task failed
    - *when*
- **fail** The task did not succeed, and the loop should skip all remaning tasks
  in the sequence and restart the loop with the next iteration. Analogous to a
  continue statement in a for loop.
    - *when*
