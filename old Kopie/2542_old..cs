using System;
using System.Reflection;
using Xunit;

public class StructsTests
{
    [Fact]
    public void IsClaimed_yes()
    {
        var ch = new ClaimsHandler();
        ch.StakeClaim(new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        ch.StakeClaim(new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        var claimed = ch.IsClaimStaked(new Plot(new Coord(1,0), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        ch.StakeClaim(new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        ch.StakeClaim(new Plot(new Coord(10,1), new Coord(20,1), new Coord(10,2), new Coord(20,2)));
        var lastClaim = ch.IsLastClaim(new Plot(new Coord(10,1), new Coord(20,1), new Coord(10,2), new Coord(20,2)));
        ch.StakeClaim(new Plot(new Coord(10,1), new Coord(20,1), new Coord(10,2), new Coord(20,2)));
        ch.StakeClaim(new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        var lastClaim = ch.IsLastClaim(new Plot(new Coord(10,1), new Coord(20,1), new Coord(10,2), new Coord(20,2)));
        var longer = new Plot(new Coord(10,1), new Coord(20,1), new Coord(10,2), new Coord(20,2));
        var shorter = new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2));
        var claimed = ch.IsClaimStaked(new Plot(new Coord(1,1), new Coord(2,1), new Coord(1,2), new Coord(2,2)));
        var claimed = ch.IsClaimStaked(CreatePlot(new Coord(1, 1), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2)));
        Assert.True(claimed);
    }

    [Fact(Skip = "Remove this Skip property to run this test")]
    public void IsClaimed_no()
    {
        var ch = new ClaimsHandler();
        ch.StakeClaim(CreatePlot(new Coord(1, 1), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2)));
        var claimed = ch.IsClaimStaked(CreatePlot(new Coord(1, 0), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2)));
        Assert.False(claimed);
    }

    [Fact(Skip = "Remove this Skip property to run this test")]
    public void IsLastClaim_yes()
    {
        var ch = new ClaimsHandler();
        ch.StakeClaim(CreatePlot(new Coord(1, 1), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2)));
        ch.StakeClaim(CreatePlot(new Coord(10, 1), new Coord(20, 1), new Coord(10, 2), new Coord(20, 2)));
        var lastClaim = ch.IsLastClaim(CreatePlot(new Coord(10, 1), new Coord(20, 1), new Coord(10, 2), new Coord(20, 2)));
        Assert.True(lastClaim);
    }

    [Fact(Skip = "Remove this Skip property to run this test")]
    public void IsLastClaim_no()
    {
        var ch = new ClaimsHandler();
        ch.StakeClaim(CreatePlot(new Coord(10, 1), new Coord(20, 1), new Coord(10, 2), new Coord(20, 2)));
        ch.StakeClaim(CreatePlot(new Coord(1, 1), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2)));
        var lastClaim = ch.IsLastClaim(CreatePlot(new Coord(10, 1), new Coord(20, 1), new Coord(10, 2), new Coord(20, 2)));
        Assert.False(lastClaim);
    }

    [Fact(Skip = "Remove this Skip property to run this test")]
    public void GetLongestSide()
    {
        var ch = new ClaimsHandler();
        var longer = CreatePlot(new Coord(10, 1), new Coord(20, 1), new Coord(10, 2), new Coord(20, 2));
        var shorter = CreatePlot(new Coord(1, 1), new Coord(2, 1), new Coord(1, 2), new Coord(2, 2));
        ch.StakeClaim(longer);
        ch.StakeClaim(shorter);
        Assert.Equal(longer, ch.GetClaimWithLongestSide());
    }

    private Plot CreatePlot(Coord coord1, Coord coord2, Coord coord3, Coord coord4)
    {
        Type plotType = typeof(Plot);
        Type[] types = new Type[] { typeof(Coord), typeof(Coord), typeof(Coord), typeof(Coord) };
        ConstructorInfo constructorInfoObj = plotType.GetConstructor(types);
        if (constructorInfoObj != null)
        {
            return (Plot)constructorInfoObj.Invoke(new object[] { coord1, coord2, coord3, coord4 });
        }

        else
        {
            throw new InvalidOperationException("Cannot find constructor that takes 4 coordinates");
        }
    }
}
