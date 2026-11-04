# Number input

Bold tabular digits in Input's well, with a raised − / + stepper on its
right. The buttons, Up and Down step by `Step` (ten steps with Shift),
within `Min` and `Max`; a button held steps once, then on and on, faster
and faster.

<!-- preview number-input 230 -->

```go
mizan.Field(c, mizan.FieldOptions{Label: "Price", Error: app.priceErr}, func() {
	mizan.NumberInput(c.Key("price"), &app.price, mizan.NumberInputOptions{Label: "Price", Min: 0, Max: 1000, Step: 0.5, Prefix: "$", Error: &app.priceErr})
})
```

Only a number can be typed: digits, one decimal point (a comma is taken as
one) with no more decimals than `Step` has, and a minus when `Min` is below
zero. What is typed applies as soon as it is a number in range on a step;
anything else marks the field invalid and says why in `*Error`, then is
settled as the field loses the focus or Enter is pressed: into range, to
the nearest step, or back to the value. Type 20 into Quantity to see it.
