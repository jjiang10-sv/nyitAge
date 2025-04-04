---Design notes---
The basis of the algorithm is, as mentioned in the assignment description, that
each gate/component in the simulation runs as its own go routine.  In general,
each component reads from its inputs, performs some logic on the inputs, and
sends to the output channels.  The biggest challenges in the design were
handling fan-out, and clock synchronization.  

For most components, I handle fan-out by making each component's output channels
buffered with a capacity equal to the number of inputs it is fanning out to.
Using buffered channels for this does create a race condition in that a single
component could potentially take multiple copies of the value from the output
channel before other fan-out targets get one.  However, in my design this is not
an issue for non-clock components, because they operate by repeatedly sending
their outputs in an infinite loop.  So even if the above scenario occurred, the
fan out targets that did not get a value in that loop iteration would block in
FIFO order and get values the next time the source component fills the output
buffers.  There is no observable difference in behavior for non clock components
in this case because they are always repeating the same algorithm in a tight
loop and absolute synchronization between them doesn't matter.

For the fan-out of clock signals, I could not use the above strategy with
buffered channels, because the scenario described (in this case, one clock
signal recipient "taking" more than one signal value from the buffer) would
result in a loss of clock synchronization between components connected to the
clock.  So, in my implementation a clock maintains an array of unbuffered
channels equal to the number of clock-driven components it is connected to, and
for each clock edge change, loops through its recipients and sends the value
over the blocking channels to each.  This ensures that components connected to
the clock remain synchronized.

Addressing another consideration regarding clock design, each component that
is driven by a clock signal (only D flip flops right now) always blocks on its
clock signal input before doing anything, so in this way it may only ever
operate in the context of one clock edge change at a time.  Finally, there is
the issue of making sure that the results of one clock edge change have
propagated all the way through the circuit before sending another clock pulse.
I solved this issue in the following way:  the main routine has a list of
"terminal" components, or in other words components that have an output that
will be read as a result of a computation.  It is fair to say that when all of
the terminal components have sent their results, then circuit propagation is
complete.

Following this logic, a clock component has an additional blocking channel for
synchronization.  After sending one clock edge change, it always blocks on this
channel until the main routine has received all results from the circuit and
unblocks it.  At this time, the clock waits for a certain amount of time based
on the configured frequency, and sends the next pulse.  This results in a slight
discrepancy between the configured clock frequency and the real time frequency
at which it operates, but this is acceptable because the "lag time" is simply a
result of the simulation execution itself, and the time is likely negligible
anyway.

---Circuit specification language (CHDL)---
In order to describe circuits in the simulator, I designed an extremely
rudimentary file format language which I dubbed Crappy Hardware Design Language,
because it's not very good and was made in a hurry. If I ever go back to this
project this is one of many things I would improve upon.  But alas, it works,
and the following are some notes on its specification:

- Comments are supported.  Any line starting with the / character will be
  completely ignored.

- Each component in the circuit gets its own line.  After it is declared, you
  will refer to this component by its line's zero-based index in the file, counting
  from the top of the file and ignoring comment lines.  From here on I will
  refer to this as the IDX of a component.

- Component lines have the following general format, where [] tokens are
  optional and <> tokens are required:
    <type> [name|value] out <[res] | <componentIDX> <componentInputIDX>> ... [out ...]
    
    Where:
    <type> = one of [and, or, nor, xor, not, nand, dff, source, clk].  This is the type
    of the component.  Most are self explanatory, but some are special:

        source - An external input to the circuit.  Has one output that
        continuously sends the specified value.  This component type has an
        additional first field in the specification: [name|value].  If
        this field is a string, the value will be taken as a name for the
        source, and the user will be prompted for its value at runtime.  If it is
        value 0 or 1, this value will be used for the source.

        dff - A D type flip flop.
        Input
            0 - Clock signal input
            1 - D input
        Output
            0 - Q output
            1 - !Q output
        This is the other component that supports the [name|value] field.  It
        works the same as the source except that it represents the initial value
        of the flip flop

        clk - A clock.  Sends pulses to its output targets at a specified
        frequency that is determined at runtime. Currently only one clock at a
        time is fully supported.  Don't add more. Or else.
        Output
            0 - Clock signal output


    [name|constant value] = This is only supported by the source and dff
    component types, see their notes above.

    out = This lets the parser know you are about to specify connections for a
    new output.  Outputs must be specified in increasing order.

    <[res] | <componentIDX> <componentInputIDX>> = Specify the IDX of a
    component you want to connect the current output to, followed by the input
    index of that component.  Alternatively, the "res" token indicates you want
    a copy of this output to be read as a result of the circuit.  You may
    specify as many of these connections as you like for each output, writing
    "out" if you want to move to the next output.

- Known issues:
    * I have no idea what happens with blank lines, but probably nothing good

    * There is nothing to check if you neglected to wire a certain input of a
      component, you can get null pointer crashes in this case

    * This language spec is almost as not very good as the language.  When in
      doubt, look at the examples.



---Example run outputs---

Enter the path to your circuit file:
8bitcounter.chdl
Parsed file successfully.
Found:
Components: 58
Terminals: 9
[] Clock frequency: 10
[] # of clock pulses to run: 30
[Output] 00000001
[Output] 00000010
[Output] 00000011
[Output] 00000100
[Output] 00000101
[Output] 00000110
[Output] 00000111
[Output] 00001000
[Output] 00001001
[Output] 00001010
[Output] 00001011
[Output] 00001100
[Output] 00001101
[Output] 00001110
[Output] 00001111
[Output] 00010000
[Output] 00010001
[Output] 00010010
[Output] 00010011
[Output] 00010100
[Output] 00010101
[Output] 00010110
[Output] 00010111
[Output] 00011000
[Output] 00011001
[Output] 00011010
[Output] 00011011
[Output] 00011100
[Output] 00011101
[Output] 00011110
[Output] 00011111


// Note that here we are specifying the numbers to be added per bit, from MSB to
// LSB.  So the input names are in the form <input><bit>.  In this example we
// are adding 0111 and 0110

Enter the path to your circuit file:
fourbitadder.chdl
Parsed file successfully.
Found:
Components: 29
Terminals: 5
[A3] Source value: 0
[B3] Source value: 0
[A2] Source value: 1
[B2] Source value: 1
[A1] Source value: 1
[B1] Source value: 1
[A0] Source value: 1
[B0] Source value: 0
[Cin] Source value: 0
[Output] 00001101
