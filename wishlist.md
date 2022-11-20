# To-do/Wishlist/Food For Thought

Here are some problems I have considered, but not implemented any solutions for as I considered them out of scope for
this project. Many of them would require additional syntax, changes to operator precedence and overall a steep increase
in complexity, bringing us closer to a scripting language than a simple definition language.


Escape characters. If you absolutely need to output << or //, you can escape these with < << < and / << /. There is
currently no way to include literal [ | ] { } _ ^ in the output. However, the intended use is natural languages, not
faux source code generation, so this is really a minor concern.


The exclusive prefix * is currently only allowed in substitutions. A more powerful implementation should allow its use
even on inline branches:

> magic [ Your lucky number is *[1 | 2 | 3 | 4 | 5]. ]


Concatenation with << is sloppy and sometimes short-circuits in undesired ways, especially in conjuction with _.

> mood [ I'm [_ | un]<<happy ].

This will not yield the desired "I'm happy" / "I'm unhappy", but crap out and say "I'mhappy". More solid implementations
of <<, _ and ^ would need to work on a node level, not as naive substitutions. However, this might actually be somewhere
between a bug and a feature. In this trivial example the workaround is simply to spell out the full [happy|unhappy]. For
other usecases (e.g. generating unintelligible fantasy names) it might actually be desirable to join syllables in
unpredictable ways.


Probability balancing. Branches have a "1 in n" probability of being chosen. This is fair but not always ideal,
especially when there are many nested groups and randomization will favor shallow top-level branches over more detailed
ones. The workaround to balance this is by manually flattening complex branches or adding multiple branches with
identical output, which both lead to error-prone data duplication.


On a related note: list concatenation. Consider this example:

> low    [ 1 | 2 ]
> high   [ 3 | 4 | 5 | 6 ]
> number [ {low} | {high} ]

This would have a 50% chance of picking 1-2 (25% chance for each), and a 50% chance of 3-6 (12.5% chance for each). This
is fine if we want 1-6 and prefer lower numbers, but to get an even distribution we would need a group that explicitly
lists 1 through 6. If we at the same time want to use low and high for other purposes this quickly becomes a frustrating
copy-paste exercise.

I propose a "union" operator that would join multiple groups together as if they had been written as one:

> number [ low && high ]

This would effectively evaluate as [ 1 | 2 | 3 | 4 | 5 | 6 ]. This could be implemented as a sub-case of the
substitution syntax or as an entirely separate operation. The order of evaluation matters here, as it should be possible
to determine the groups to be joined through random selection:

> group1 [ group2 && [group3 | group4] ]

Additional difficulty will arise when this is combined with the exclusive * prefix.


Caching. It would be useful if values could be captured and reused:

> introduction [ His name was {name=[ Eero | Alvar | Jari ]}. {name} was his name. ]

Some concerns are syntax, scoping and mutability. Are variables local to the group they are declared in, or should they
be global?  What if a variable is used uninitialized? What if a variable is overwritten? Should they always be evaluated
depth-first?


Tuples. It would be useful if values could be grouped by context. This is really a must-have if we want to create
natural language fiction with any degree of consistency. Consider something like this:

> bad_story [ My [brother|sister] called to tell me that [he|she] had lost [his|her] phone. ]

Here we have no interdependency between the branches and the gender of the sibling will be mixed up. A trivial
implementation could allow tagging groups as "friends", so that the first branch chosen would enforce the same relative
branching in the others. Unfortunately this logic will quickly break down for groups of unmatched sizes or when we want
to use substitutions. It is also difficult to read and maintain multiple uncorrelated branches in parallell. A more
ambitious approach could be using maps (where one value is used to retrieve another value), or a construct that allows
caching grouped values (this will also escalate in complexity quickly):

> bad_story
> [
>   (sibling, they, their) = [ (brother, he, his) | (sister, she, her) ]
>
>   My {sibling} called to tell me that {they} had lost {their} phone. ]
> ]

A termination operator that, when encountered, speedily finalizes evaluation of the current identifier, without
including anything from further down the tree.
