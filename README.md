Simple Enumish Generator For Go
===============================

Add string or int enums to your project:

    package mystuff

    type MyString string

    const (
        Foo MyString = "foo"
        Bar MyString = "bar"
    )

    type MyInt int

    const (
        Inty MyInt = 1
        Squinty MyInt = 2
    )

Add `tools.go` to your project as per [#25922](https://github.com/golang/go/issues/25922),
unless that guidance has changed recently, which is entirely possible. This has nothing
to do with enums or this project, it's just how you tell go modules about this tool:

    //+build tools

    package tools

    import _ "github.com/shabbyrobe/go-enumgen/cmd/enumgen"

Add you some go generate into the same package as your enums (not in
`tools.go`, which is just a bit of "go modules administrativa"):

    //go:generate enumgen MyString MyInt

Run you some go generate inside the package with your enums:

    $ go generate

And now you can do fun stuff:

    var x = Foo
    fmt.Println(x.Name()) // "Foo"

    var nup MyString = MyString("INVALID")
    fmt.Println(nup.IsValid()) // "false"

    var fs = flag.NewFlagSet()
    var v Foo
    fs.Var(&v, "mystring", "Set one of these generated things from a flag")

