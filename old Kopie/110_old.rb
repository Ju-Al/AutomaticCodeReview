<div class="contributing">
  <div class="container">
    <div id="intro">
      <h1>How to Contribute</h1>
      <p>
        <strong>The Odin Project</strong> is crowdsourced.  It is built both by experienced engineers
        and by people who are currently learning to code
        <em>just like you</em>. Regardless of your experience, you can contribute... in fact,
        we hope you do! Even if you
        don't feel ready, you can still get involved and observe until you're comfortable. Still
        not sure? Here are some <a id="reasonsToggle" href="#reasonsModal" data-toggle = "modal">reasons to get
        involved!</a>
      </p>
      <div id="reasonsModal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="reasonsModalLabel" aria-hidden="true">
        <div class="modal-header">
          <button type="button" class="close" data-dismiss="modal" aria-hidden="true">x</button>
          <h3 id="reasonsModalLabel">Reasons to get involved</h3>
        </div>

        <div class="modal-body">
          <p >The Odin Project is designed around a simple idea- make it easy for you to
            get started coding! We want to make it easy for newcomers to get into an
            <em>actual</em> real-world coding environment so you can get the experience
            you need. <strong>Here's the benefits to you:</strong>
          </p>
          <ul>
            <li><span class='highlight'>Flexibility--</span> You can work on your own time. It's not 9 to 5 so you can
              get involved when it's convenient for you.</li>
            <li><span class="highlight">Exposure--</span> Contributors to The Odin Project will be listed on the page
              and on Github, where you can show people your work!</li>
            <li><span class="highlight">Relevant Training--</span> This is the development that is in demand and
              happening today.</li>
            <li><span class="highlight">Job Focus--</span> No matter why you are learning to code, the right skills matter.
              You'll learn good habits and discipline that will put you at the top.</li>
            <li><span class="highlight">Community--</span> Meet like-minded people and start building your network now.
              Find others that can work in your schedule and interests.</li>
          </ul>
        </div>

        <div class="modal-footer">
          <h3>But mostly... it's fun!</h3>
        </div>
      </div>
      <p class="light-emphasis"><em>
        This guide will help you get started contributing to the Odin Project website (which you're on now!).
        If you're looking to contribute to the <strong>**curriculum**</strong> (for instance to add your solution to one of the projects),
        <a href="https://github.com/TheOdinProject/curriculum/blob/master/contributing.md">check out the
        instructions posted here</a> instead.
      </em></p>

      <p>
        The roles below are volunteer positions.  Badges have no cash value but may be redeemed for hugs at our discretion.
      </p>

      <h2>Where do I start?</h2>
    </div>

    <div id="newbie" class="container">
      <div class="row-fluid skill-header">
        <div class="span2 star">
          <%=image_tag "icons/star_newbie_sm.png", title: "Advanced Beginner"%>
        </div>
        <div class="span10">
          <h2>As a Total Newbie...</h2>
          <p>
            If you're just learning to program, this is where to start. You can still help out! Click below to see
            how you can get involved and then check out the Contributor's Hall of Fame at the bottom of the
            page. That's your goal!
          </p>
        </div>
        <div class="text-center">
          <a id='newbie-path-link' class="more-button" data-toggle="collapse" data-target="#newbie-path">More</a>
        </div>
      </div>

      <div id="newbie-path" class="row-fluid collapse">
        <div id="newbie-path-1" class="paths">
          <h3>Student/Observer</h3>
          <p>
            Your first step as a newbie should be to check out <%= link_to "the curriculum", courses_path %>.
            The learning curve can be steep at first -- <strong>but don't cut corners!</strong>
            While you're working through the lessons there, you should start getting your feet wet
            by observing how this website is being built -- it will give you great context for
            what you're learning.
          </p>

          <h4>How to Get Started</h4>
          <ol>
            <li>
              Start working through the <%= link_to "curriculum", courses_path %> on your own
            </li>
            <li>
              Check out the <%= link_to "student study group", studygroup_path %> for support
            </li>
            <li>
              Complete part 1 of the <%= link_to "On-Boarding", "#onboarding" %> section below
            </li>
            <li>
              Observe our group pairing sessions (<%= link_to "see the Community for details", "https://plus.google.com/communities/100013596437379837846" %>)
            </li>
            <li>
              Write a lot of code!
            </li>
          </ol>
          <div>
            <h4>The Next Step</h4>
            <ul>
              <li>
                Graduate to <span class="highlight">Tester/Bug Sniper</span> when you're ready!
              </li>
            </ul>
          </div>
        </div>

        <div id="newbie-path-2" class="paths">
          <h3>Tester/Bug Sniper</h3>
          <p>
            The most common first step for getting actively involved with a new code base is to fix a bug,
            add tests to an area that is weak, or just improve comments and documentation (like this page).
            Put your engineer hat on and jump into the project!
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Complete part 1 and 2 of the <%= link_to "On-Boarding", "#onboarding" %> section below.
            </li>
            <li>
              Try to catch a bug as you navigate the site and then figure out what is wrong.
              Post it (with your solution if you have one) in the <%= link_to "community", "https://plus.google.com/communities/100013596437379837846" %>
              and say hi in one of our regular SCRUM meetings.
            </li>
            <li>
              Sit in on pairing sessions or <%= link_to "watch the videos", "https://www.youtube.com/channel/UCk0b0VTnJxXbxupXJ4D2yjQ"%> to identify areas that need help.
            </li>
            <li>
              Keep working through the <%= link_to "curriculum", curriculum_path %>!
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Bug Zapper</span>, <span class="highlight">Documentarian</span> or
              <span class="highlight">Test Dummy</span> badges by having a pull request accepted.
            </li>
            <li>
              Graduate to the <span class="highlight">Advanced Beginner Skill Level</span> once you've completed the
              <%= link_to "Ruby on Rails Course", lessons_path("ruby-on-rails") %> and
              had a pull request accepted.
            </li>
          </ul>
        </div>
      </div>
    </div>

		<div id="advanced-beginner" class="container">
		  <div class="row-fluid skill-header">
		    <div class="span2 star">
		      <%=image_tag "icons/star_adv_beginner_sm.png", title: "Advanced Beginner"%>
		    </div>
		    <div class="span10">
		      <h2>As an Advanced Beginner...</h2>
          <p>
            As an Advanced Beginner, you have some experience -- probably a few classes,
            books, and tutorials under your belt. You have gaps in your knowledge
            but you know the fundamentals. You should keep working on completing the courses you're
            unfamiliar with (especially <%= link_to "Rails", lessons_path("ruby-on-rails") %>),
            but it's time to consider taking an active role in the project to really start applying
            what you've learned.
          </p>
        </div>
        <div class="text-center">
          <a id = "adv-beginner-path-link" class="more-button" data-toggle="collapse" data-target="#adv-beginner-path">More</a>
        </div>
      </div>

      <div id="adv-beginner-path" class="row-fluid collapse">
        <div id="adv-beginner-path-1" class="paths">
          <h3>Contributor</h3>
          <p>
            Start here! Join the team and start collaborating on existing stories during our
            group coding sessions.  We coordinate everything through our regular SCRUM meetings
            (posted to the <%= link_to "Community", "https://plus.google.com/communities/100013596437379837846" %>)
            so be sure to swing by and see what you can work on.
            You may not feel ready, but give it a shot! We're easy to work with.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Keep working through the <%= link_to "curriculum", courses_path %> as needed to get your skills up.
            </li>
            <li>
              Complete the <%= link_to "On-Boarding", "#onboarding" %> section below if you haven't already.
            </li>
            <li>
              Attend the SCRUM meetings and start pair- or group-programming on actual stories with other
              team members.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Contributor</span> badge by making a meaningful contribution
              to a story.  Earn your <span class="highlight">Committer</span> badge by have a pull request
              accepted for a small individual story.
            </li>
            <li>
              Graduate to the <span class="highlight">Story Owner Role</span> once you've contributed to
              three stories or earned your <span class="highlight">Committer</span> badge.
            </li>
          </ul>
        </div>

        <div id="adv-beginner-path-2" class="paths">
          <h3>Story Owner</h3>
          <p>
            The Story Owner takes ownership of a story in Pivotal Tracker! It's up to you to see it through,
            present it to the team, and get it accepted!  You'll also need to set up the story/feature branch(es)
            and coordinate any group coding sessions around the story.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Complete the <%= link_to "On-Boarding", "#onboarding" %> section if you haven't yet.
            </li>
            <li>
              Attend a SCRUM meeting and volunteer for a story you'd like to complete.
            </li>
            <li>
              Push the 'Start' button in <%= link_to "Pivotal Tracker", "https://www.pivotaltracker.com/s/projects/979092" %>
              and get going!
            </li>
            <li>
              Coordinate group coding sessions with other contributors via the
              <%= link_to "Community", "https://plus.google.com/communities/100013596437379837846" %>.
            </li>
            <li>
              When you've completed the story, submit a pull request and present it at the next SCRUM!
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Story Owner</span> badge by having pull requests
              accepted for three stories.  Earn your <span class="highlight">Feature Owner</span>
              badge by have a pull request accepted for a feature involving multiple individual stories.
            </li>
            <li>
              Graduate to the <span class="highlight">Project Manager Role</span> once you've earned both
              your <span class="highlight">Story Owner</span> and <span class="highlight">Feature Owner</span> badges.
            </li>
          </ul>
        </div>

        <div id="adv-beginner-path-3" class="paths">
          <h3>Project Manager</h3>
          <p>
            By now you should be thoroughly familiar with the project. Your mission:
            lead a team to complete an epic or series of stories and features.  This could mean
            organizing coding sessions, finding knowledge team members need to move on, or
            rolling your sleeves up and coding... basically anything it takes to get the story delivered!
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Strengthen your understanding of software engineering. Consider taking the
              169.x courses on <%= link_to "Edx.", "https://www.edx.org/course-list/allschools/computer-science/allcourses" %>
            </li>
            <li>
              Learn more about Project Management and SCRUM Mastering from these (TBA) resources.
            </li>
            <li>
              Take on responsibility for delivering an Epic (see <%= link_to "Pivotal Tracker", "https://www.pivotaltracker.com/s/projects/979092" %>).
            </li>
            <li>
              Lead daily or weekly SCRUM meetings organizing project development.
            </li>
            <li>
              <span>Graduate!</span> Go to Experienced Engineer skill level!
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Project Manager</span> badge by delivering your epic / featureset
              or working continuously with the project in a PM role for 6 weeks.
            </li>
            <li>
              You can continue to gain experience working in the <span class="highlight">Project Manager Role</span>. Start
              looking into the <span class="highlight">Experienced Engineer</span> paths for your next challenge when you
              feel ready.
            </li>
          </ul>
        </div>
      </div>
		</div>

		<div id="experienced-engineer" class="container">
		  <div class="row-fluid skill-header">
		    <div class="span2 star">
		      <%=image_tag "icons/star_engineer_sm.png", title: "Experienced Engineer"%>
		    </div>
		    <div class="span10">
		      <h2>As an Experienced Engineer...</h2>
          <p>
            We've got several very important ways that experienced engineers can help...
            your coding and leadership skills help drive the project and guide others as they learn.
            It's a most welcome way to give back to the community.
          </p>
        </div>
        <div class="text-center">
          <a id = "engineer-path-link" class="more-button" data-toggle="collapse" data-target="#engineer-path">More</a>
        </div>
      </div>

      <div id="engineer-path" class="row-fluid collapse">
        <div id="engineer-path-1" class="paths">
          <h3>Mentor</h3>
          <p>
            This is a project primarily driven by the efforts of new developers.
            That means you can help out significantly just by occasionally uncrossing wires and gently
            administering best practices.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Complete the <%= link_to "On-Boarding", "#onboarding" %> section if you haven't yet.
            </li>
            <li>
              Help guide other coders during group coding sessions (
              <%= link_to "see the Community", "https://plus.google.com/communities/100013596437379837846" %>
              and attend any regular SCRUM session for details).
            </li>
            <li>
              Perform code reviews for pull requests -- which is extremely helpful for beginners!
            </li>
            <li>
              Demonstrate best practices by taking on any of the roles in the Advanced Beginner section above.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Mentor</span> badge by
              working with students of the project
              for 2 or more hours per week over 6 weeks.
            </li>
            <li>
              Check out the <span class="highlight">Project Advisor Role</span>
              if you'd like to be more
              involved at a high level or the
              <span class="highlight">Code Warrior Role</span> if you'd
              like to get your hands dirty with more code.
            </li>
          </ul>
        </div>

        <div id="engineer-path-2" class="paths">
          <h3>Code Warrior</h3>
          <p>
            Sometimes you want to help out but prefer to stick with building stuff.
            In that case, you're welcome to pick up user stories from the backlog or
            work with project organizers for a more custom role.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Take the lead on some of the harder stories with the help of other team members.
              Even if you'll be solo-ing, please let more junior coders watch!
            </li>
            <li>
              Suggest tools and methods that are better than currently implemented solutions.
            </li>
            <li>
              Help others to prioritize development of their skillsets, based on your experience.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Code Warrior</span> badge by sustaining a high level
              of development output over 4 or more weeks.
            </li>
            <li>
              Check out the <span class="highlight">Project Advisor Role</span>
              if you'd like to be more
              involved at a high level or the <span class="highlight">Mentor Role</span>
              if the idea of doing a bit of teaching is starting to rub off on you.
            </li>
          </ul>
        </div>

        <div id="engineer-path-3" class="paths">
          <h3>Project Advisor</h3>
          <p>
            A Project Advisor takes a more ongoing and high level view than a typical
            mentor -- your involvement is at the project level.  You will help with
            everything from project architecture to major refactorings and implementation
            of industry best practices.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Work with project leadership to implement best practices and processes.
            </li>
            <li>
              Help identify code smells and brittle code that needs refactoring.  Work
              with the team to refactor it.
            </li>
            <li>
              Create stories that less experienced members can tackle and finish.
            </li>
            <li>
              Contribute to overall project strategy.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Advisor</span> badge by staying involved
              with the project (> 2hrs/wk) as an advisor over three or more months.
            </li>
            <li>
              Check out the <span class="highlight">Code Warrior Role</span> if you'd
              like to get your hands dirty with more code.
            </li>
          </ul>
        </div>
      </div>
		</div>

		<div id="nontechnical" class="container">
		  <div class="row-fluid skill-header">
		    <div class="span2 star">
		      <%=image_tag "icons/star_non_technical_sm.png", title: "Beginner"%>
		    </div>
		    <div class="span10">
		      <h2>As a Non-Developer (or Willing Intern)...</h2>
          <p>
            The scope of this project spans much more than just code -- we can benefit from help
            with everything from graphic design to user experience (UX) to community management.
            We believe firmly in the power of a well designed student experience and a loyal community
            of students and contributors. While certain technical
            skills are still required, these roles emphasize creativity,
            intuitiveness, and imagination.
          </p>
        </div>
        <div class="text-center">
          <a id = "nontechnical-path-link" class="more-button" data-toggle="collapse" data-target="#nontechnical-path">More</a>
        </div>
      </div>

      <div id="nontechnical-path" class="row-fluid collapse">
        <div id="nontechnical-path-1" class="paths">
          <h3>Designer/UX</h3>
          <p>
            We're constantly trying to come up with more engaging student
            experiences as well as a more effective information delivery
            and content organization strategy.  The look and feel of the
            site will soon be under major overhaul, so your design skills
            will help greatly.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Complete part 1 of the <%= link_to "On-Boarding", "#onboarding" %>
              section if you haven't yet.
            </li>
            <li>
              Attend the SCRUM meetings and get involved in the <%= link_to "community", "https://plus.google.com/communities/100013596437379837846" %>
              to see where you can contribute.
            </li>
            <li>
              Provide critical feedback on the site and suggestions
              for improvement.  Help students and contributors improve their design intuition.
            </li>
            <li>
              Design user flows and graphical assets that enhance the
              student experience.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Designer</span> badge by delivering
              creative assets, participating in a UX flows redesign, or
              staying involved
              with the project (> 2hrs/wk) as a creative advisor over
              six or more weeks.
            </li>
          </ul>
        </div>

        <div id="nontechnical-path-2" class="paths">
          <h3>Marketing & Community</h3>
          <p>
            The power of this project comes from its community of students
            and contributors.  If you're able to help us get the word out
            and keep our community happy and engaged and learning, your
            contributions are most welcome.
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Complete part 1 of the <%= link_to "On-Boarding", "#onboarding" %>
              section if you haven't yet.
            </li>
            <li>
              Attend the regular SCRUM meetings and get involved
              in the <%= link_to "community", "https://plus.google.com/communities/100013596437379837846" %> to see where you can contribute.
            </li>
            <li>
              Spread the word about the project and advise
              us on how we can improve its presentation.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Community</span> badge by
              actively fostering community development over several months.
              Earn your <span class="highlight">Market Master</span> badge
              by being actively involved in the marketing of the project over
              several months.
            </li>
          </ul>
        </div>

        <div id="nontechnical-path-3" class="paths">
          <h3>Project Intern</h3>
          <p>
            There's a whole lot of work involved with keeping this project
            running smoothly and moving forward.  If you'd like to reach into
            the guts and help drive from the inside, consider volunteering
            as an intern.  Your role will span every other role mentioned on here
            so you'd better be ready for it!
          </p>
          <h4>How to Get Started</h4>
          <ol>
            <li>
              Manage the feature development pipeline in the context of
              overal project strategy.
            </li>
            <li>
              Do as much coding as you're capable of.
            </li>
            <li>
              Take on every other listed role and earn as many badges as
              you can.
            </li>
          </ol>
          <h4>The Next Step</h4>
          <ul>
            <li>
              Earn your <span class="highlight">Odintern</span> badge by working
              your butt off for at least eight weeks.
            </li>
          </ul>
        </div>
      </div>
		</div>

    <div id="onboarding">
      <h2>On-Boarding</h2>
      <p>
        The purpose of on-boarding is to get you up to speed on the project and how it is
        managed. You will need to familiarize yourself with the tools and websites
        used by the team. The following instructions will guide you through part 1 and part 2 of the setup:
      </p>
      <h3 class="center">Part 1: The Basics</h3>
      <div id="part-1" class="paths">
        <p>
          <span class="highlight">Basic Setup</span>
        </p>
        <ul>
          <li>
<<<<<<< HEAD
            First, read the <%= link_to "Getting Involved", getting_involved_path %> page for a detailed overview and instructions on how to proceed. Don't worry if it doesn't all make sense just yet-
=======
            First, read the <%= link_to "Getting Involved", "https://github.com/TheOdinProject/theodinproject/blob/master/getting_involved.md"%> document for a detailed overview and instructions on how to proceed. Don't worry if it doesn't all make sense just yet-
>>>>>>> e31eeffde1e60c9b968bc01dbf275e17a8fe31be
            it's enough for now to get some exposure.
          </li>
          <li>
            Set up a <%= link_to "Github", "http://github.com"%> account if you haven't yet
          </li>
          <li>
            Set up an account on <%= link_to "Pivotal Tracker", "http://pivotaltracker.com"%> and check out <%= link_to "The Odin Project", "https://www.pivotaltracker.com/s/projects/979092" %>
          </li>
          <li>
            There are a growing number of online development environments (called an IDE for <em>integrated development environment</em>) that are very easy to use to get started. The Odin project often uses <%= link_to "Nitrous.io", "https://www.nitrous.io/join/Nex2i3R7Vyc?utm_source=nitrous.io&utm_medium=copypaste&utm_campaign=referral"%>. Create an account here and try it out!
          </li>
          <li>
            Join us on our <%= link_to "Google Community", "https://plus.google.com/communities/100013596437379837846" %>
          </li>
        </ul>
      </div>

      <h3 class="center">Part 2: The Details</h3>
      <div id="part-2" class="paths">
        <p>
          <span class="highlight">Advanced Setup</span>
        </p>
        <ul>
          <li>
<<<<<<< HEAD
            Completely go through the <%= link_to "Getting Involved", getting_involved_path %> page and make sure you have a firm grasp of the contents. Ask questions in the SCRUMs or in the community if you are unclear on
=======
            Completely go through the <%= link_to "Getting Involved", "https://github.com/TheOdinProject/theodinproject/blob/master/getting_involved.md"%> document and make sure you have a firm grasp of the contents. Ask questions in the SCRUMs or in the community if you are unclear on
>>>>>>> e31eeffde1e60c9b968bc01dbf275e17a8fe31be
            anything.
          </li>
          <li>
            Fork the Odin Project <%= link_to "Repo", "https://github.com/TheOdinProject/theodinproject" %> to
            your own Github account and then clone it from there to your local machine (or Nitrous box).
          </li>
          <li>
            If working on your own story in <%= link_to "Pivotal Tracker", "http://pivotaltracker.com"%>, fork a branch
            off your local repo to work on. Give it a
            relevant name and pair with a partner if you can. You can create a matching branch on your Github
            repo and push up to it from your local branch so your partner can work on it. If you own the story,
<<<<<<< HEAD
            this is where the code will stay until you have a working feature. The <%= link_to "Getting Involved", getting_involved_path %> page contains detailed instructions.
=======
            this is where the code will stay until you have a working feature. The <%= link_to "Getting Involved", "https://github.com/TheOdinProject/theodinproject/blob/master/getting_involved.md"%> document contains detailed instructions.
>>>>>>> e31eeffde1e60c9b968bc01dbf275e17a8fe31be
          </li>
          <li>
            Many people working in open source software work on Mac or linux. On Windows, a VM (virtual
            machine) running Linux is probably best. Setting up a VM running Linux and learning some basic
            BASH command line skills will prove very useful. You can follow the instructions for setting up
            a VM running Ubuntu at the book site for the Berkeley 169.x course <%= link_to "here", "http://beta.saasbook.info/bookware-vm-instructions"%>. (The course is highly recommended once you have
            gotten up to speed, see 'Advanced Beginner' section above).
          </li>
          <li>
            For detailed instructions on setting up various environments, you can also see the <%= link_to "project readme", "https://github.com/TheOdinProject/theodinproject/blob/master/README.md" %>.
          </li>
          <li>
            Join us on our <%= link_to "Google Community", "https://plus.google.com/communities/100013596437379837846" %> and start contributing!
          </li>
        </ul>
      </div>
    </div>

    <div>
		  <h2>CONTACT</h2>
		  <p>
		    Check out our <%= link_to "Google Community", "https://plus.google.com/communities/100013596437379837846" %>
        for discussions and events around those efforts, as well
        as additional links for learning more about this living and
        growing adventure we call The Odin Project.  Get in touch via
        the <%= link_to "Contact Page", contact_path %>.
		  </p>
		</div>
    
    <%= render "contributors" %>
    
  </div>
</div>
