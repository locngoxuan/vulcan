# Vulcan Executor

In the beginning, Vulcan create shell script for executing build steps on each job. The limitation of this approach is difficult to save output of each steps for using in other steps afterwards. Then Vulcan Executor is created. Goals of this executor is able to control each build step and manipulates input/output of every steps.