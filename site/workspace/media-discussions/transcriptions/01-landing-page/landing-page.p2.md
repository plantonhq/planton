Swarup Donepudi  0:00  
Okay, going back. So good swap, build, deploy, deploy, deploy. Build will have a conversation with plantor saying, Hey, can you add GitHub actions to my repository to set up build automation for building and, yeah, the same things that we talked about. Then deploy, then it will conversation. Hey, I want to deploy my service. I will ask, where is your repository? And then it says, Would you like to deploy to Kubernetes? It will say, yes, okay, what is your name of the service? And it will show the micro service thing and, or maybe it can be as simple as, okay, I have added GitHub a pull request to add this thing, and the pull request screenshot will show customize and something like that, right? Okay, so, so far, Bootstrap, build and deploy. Yeah, these three

Speaker 1  0:53  
commanders service category of building and deploy, right? Yeah, provisioning. There is nothing we are talking about provisioning. By provisioning, I mean

Swarup Donepudi  1:05  
infrastructure that, yeah, that we require true to build. So this is now we are getting into the the domain of internal development platforms, where the idea is, there is platform engineering group, and then there is developer group. So platform engineering group does a certain set of things, and developers do certain things, but all of them, the end goal is to make sure the black whatever platform engineering does, they'll enable self service for developers. Okay, so what is it that platform engineering teams do within the companies is allow developers to bootstrap services, deploy monitor, set up monitoring, or whatever right operate so they can do that all by themselves. Platform engineering, on the other hand, will make sure they are setting up infra they are sharing, like allowing the environments to go use those info. So, so maybe we can have two journeys as our own, like unique proposition, whether you are platform engineer or whether you are a developer, if you are a platform engineer, then provision and probably share or configure, I don't know, or do you want a single train here, like provision?

Suresh Attaluri  2:35  
Yeah, provision, code, provision,

Swarup Donepudi  2:37  
bootstrap

Speaker 1  2:39  
provision is the first thing that happens, right? Yeah. Provision, build, deploy, sorry. Provision, Bootstrap, build and deploy. Yeah, I'll write those up here. Let's write all those

Swarup Donepudi  2:58  
now. Those will be created by chatgpt automatically

Speaker 1  3:00  
later, we can filter few if we can interact. Yeah,

Swarup Donepudi  3:05  
so, provision, Bootstrap, build,

Suresh Attaluri  3:12  
deploy, deploy.

Swarup Donepudi  3:17  
So here we have, I mean, our platforms, capabilities. You can deploy services, you can deploy open source, you can deploy managed services. So this is where multi cloud deployments come in, right?

Speaker 1  3:32  
So in that case, provision and this deploy are same,

Swarup Donepudi  3:36  
no provision has a different meaning. So you need for these services or open source to land, you're creating some infrastructure. So this is provisioning infrastructure, whereas these deployments, right? What is that that is done on a day to day basis, is for these services, they need open source software, and developers would be in management services like s3 buckets or whatever. So this stage of deployment is only talking about deploying whatever is needed for services, and the provisioning step is talking about where the deployment setting up infra. So infra setting up in first environments. So yeah. In fact, this goes back the same story. Goes back to December 2022 or January 2023 when we submitted our first application to Y Combinator and we told the exact same story. The first step was provisioning a Kubernetes cluster. Second step was creating an environment. Third step was bootstrapping the repository. So we demonstrated all of that in like the 20 minute recording, right? And then we automated build which we called ahead magic pipeline, and which also confidence deployment. And then we did the day two operations, which is users were able to list parts and stream logs for I think just streaming logs was the only thing we did in our first version. But the story of narrative did change. It's just the implementation change over time. So yeah, we will see again. We don't have to, probably, like, nail down this thing, this journey, because this is all of this is being recorded as a conversation. We can also rely on chatgpt to help us understand what would make a proper journey. It would do a comparison of what is on GitHub, which is coding, planning, collaborating, automating and securing, and it will understand what plant on Cloud is about based on all of this conversation, and try to come up with its own set. And we get, we can do a comparison, yeah, go on.

Speaker 1  5:47  
So I think, yeah, provision, provision is deploying GK cluster at all this,

Swarup Donepudi  5:55  
yeah, maybe in deploy step itself, we can show, I want to also add a Redis survey, Redis to my service. And I want, I need an s3 bucket, okay, yeah, right, so, but I think those can be like screenshots here, right? Yes, we'll see this,

Speaker 1  6:19  
but I think we have clear definitions for these

Swarup Donepudi  6:22  
four, yeah. And then the next step is operate, which is what is called as day two plus operations. So here we are showing we are saying, Can you give me parts, these parts?

Speaker 1  6:39  
So this is only specific to Google, yeah,

Swarup Donepudi  6:43  
but we want to make sure, yeah, that's fine. We will probably we are not trying to win the entire world by saying we do everything right now. I think the main core selling point that we are trying to achieve with this new sales website is we want to help the visitors understand the core problem that we are solving, so which they can get with this. So we are claiming DevOps at the start, and then we are saying setting the gold standard for Internet developer platforms. And then these labels and the content here will probably either tell the visitor that it is either for them or not for them. It's my, oh, I came here to order food, but it looks like this is all related to software engineering, so he goes out, so something like that. Okay, so, so operate, yeah, we set Kubernetes pods streaming logs,

Speaker 1  7:37  
and there's action of operators to show the Kubernetes law. Or maybe we

Swarup Donepudi  7:42  
can also say something like, can you take a backup of my RDS database? So this is something, again, that is in our pipeline, an operational task, yeah, So day two plus ops itself is actually starting to take its own shape. I mean, I wrote up some kind of architecture and design around it, meaning, so we did Kubernetes data ops, right? So that seems like a foundation, which is, which is what I can think of like going back in the time the pulumi module automation enabled us, and we are calling today that as a multi cloud deployment framework. So it evolved, right? So Kubernetes integrations that are whatever we did in starting from like early days and now they stood slowly evolving to become a much bigger abstraction called day two operations. And we can call them plant and product actions. We can let developers define what those actions are. We define the framework. So there is a lot that's and even I've seen it in other platforms that they are calling, like HashiCorp waypoint. They call it as Golden workflows. Okay, so they believe that once you deploy, you want to do day two operations on what's deployed. Yes. Examples are, while Kubernetes, listing parts is one thing you want to restart. A part is another thing. So platform engineers need to develop these self service capabilities for day two and beyond. And those are not only for Kubernetes. Those can be for anything which is, can you take give me a backup for my RDS database? Can you scale my service to so and so? So these are all like day two, plus actions. So operate comes under that thing. So what we are going to claim on the website, we will get to that when we get there provision

Speaker 1  9:34  
or deploy those things, which you initially do. And there are certain operations you would like any DevOps engineer

Swarup Donepudi  9:43  
or any developer developer would needs to do for their own service beyond deployment. So there, that is where we are calling it as operate, operate and monitor. Monitor is like so deploy, monitor, operate or deploy, operate, monitor, whatever the order can be, anything monitor is like, can you add like, set up an alert for this service. Is capability that users want. So we're not doing it right now. We will eventually, but not not right now. So we can say we don't have to, because if we want to tell a story, we can say operate, but we can say coming soon. But it is 100% sure that it is it has to land on this internal developer platform, because without it, it's not a complete story. They need to go so developers are not capable of observing like setting up observability for their own app, and that's a big gap. So for Internet developer platforms, these are, like, necessary, and we are, like, adding a nice layer on top of that with conversational interfaces that is, like the product that I've been thinking about. So yeah, I think these are getting really good, like provision, Bootstrap, build and deploy, or whatever, automate and operate and monitor or observe. Monitor or monitor is good? Monitoring, observing? Yeah, we'll let chatgpt also decide some of these. So I think I'll work with you to first create this element, yes, and then we will work with on each of those to finalize the screenshots or any recording or tips will again pair with your share on that. I think that that's

Suresh Attaluri  11:36  
like,

Swarup Donepudi  11:41  
oh, okay, yeah. So the story that we are saying selling on our own website, if you go back to this thing, so features, plan Dora, which is like the conversational self service DevOps,

Speaker 1  11:57  
that's the provisioning, bootstrapping, build and deploy. Everything

Swarup Donepudi  12:01  
is self service, yeah, and then service hub, service, that is where, like you can see all the services, etc. So, okay, are we telling it in that story is discover, is what I think it's hard for us to probably put it in

Speaker 1  12:26  
integrated ISC workflows. Well, when you say integrated IAC workflows, are those infra jobs,

Swarup Donepudi  12:31  
right? Yeah, like meaning you make a configuration change or you want to deploy something, something you workflow automatically kicks kicks off. So, yeah. So

Suresh Attaluri  12:41  
can we put

Swarup Donepudi  12:45  
more customers

Speaker 1  12:48  
and highlight that have these real time logs coming up?

Swarup Donepudi  12:54  
So this one is a section title, so this one has its own media,

Suresh Attaluri  13:01  
so that is a feature, right, like

Swarup Donepudi  13:05  
so this is right. So here you will show those log streaming, pulumi plan or TerraForm.

Speaker 1  13:13  
What is that user is actually? User is getting with these logs. That's what is like with self service DevOps. User is able to provision infra. User is able to bootstrap and deploy, yeah, and operate also. But can we think of a word if we say, if we show these logs. Just if we can't, then we'll move on.

Swarup Donepudi  13:45  
Yeah, I am having a hard time understanding the exact point you are trying to highlight here, which is, so what is that? Again, it goes back to this this itself, this structure itself. Okay, so I think I told you about I wanted to begin with yesterday, begin with the structure that we wanted to come with in establishing this content, right? So we took inspiration from GitHub, right? Again, that was the fund of our success, which is take one reference and build arrow, right? So this is what. So when you say actions, can you now, can you rephrase your question, aligning that with what GitHub is

Speaker 1  14:37  
saying, right? Yeah, actions is the feature. What is that user is going to get? Is user is going to automate their workflows, right? So that's why they have mapped that actions with this automate true thing. Yeah,

Swarup Donepudi  14:52  
so in our in our case, so if you go scroll on top features, and I integrated IAC workflows, preview and pulumi, this is our sponge line, whereas their punch line is automate workflows.

Speaker 1  15:11  
So can we derive a word for IAC workflows so that we can add next to these provisions? I mean

Swarup Donepudi  15:19  
Ise workflows is comparable to actions in meetup. Like if we take things, we took we are not innovating there, we took one strong inspiration. So everything that we did, we tried to find counterparts for great thing. So

Speaker 1  15:35  
can we also use a category as automate and and demo these ISC workflow. ISC workflows, I mean, along with all these,

Swarup Donepudi  15:50  
okay, again, going back to GitHub reference. So important to note that this are not aligned with the rest of the page. So here they are telling one story, and if and now going case if you screen keep scrolling down, they are disconnected now, meaning so they they resorted to explaining each of these, but not in conjunction with this. So yeah, like So these tend to be more like their features, core features, okay, here they are telling one narrative, one complete narrative, like you can use GitHub to do code plan collaborate, automate, secure,

Speaker 1  16:36  
done, but each of these, code plan, collaborate, automate and secure is associated with a feature,

Swarup Donepudi  16:44  
right? But, but that they are not map, establishing a map between those like, for example, take,

Speaker 1  16:53  
like, Code Spaces like, I know automate is mapped with actions. Now, if you click on that also, you'll see they're talking about actions only, okay, you're talking about a feature. Okay, go on Similarly, when we look at code, we're talking about their co pilot thing, okay, plan, whether they're talking about their feature of something called issues, yeah, this one.

Swarup Donepudi  17:17  
Plan, no, that's not issues. That's a plan is not here, okay?

Speaker 1  17:32  
And collaborate. They're using those words discussions, those discussions are using it as collaborate, okay, yeah, certain features, they don't, they might not feel as important, so they haven't included,

Swarup Donepudi  17:45  
yeah, that's, that is the core point that I'm trying to establish, even in our case, which is, we, these are not directly doing this. They are related in some way. You can, we can draw like relationships, I'm pretty sure so. And I can do, I can also do the same thing for our own which is, like, we didn't, we are talking about this workflow, right? We didn't have that

Speaker 1  18:11  
workflow like this. COVID is dashboard. We can map it to operate, yeah,

Swarup Donepudi  18:15  
IAC workflows, mapping it to automate, automate.

Speaker 1  18:19  
That's the whole point. We can add it to this list, right? Automate is a new category,

Swarup Donepudi  18:26  
or we do it for both, yeah, or so we brainstorm on the possibility of combining these two to become automate. Automate. But IAC workflows itself is related to these two things in our own like system, but we are kind of splitting this journey to tell a

Speaker 1  18:44  
story. Yeah, they can be explained in these two categories, also, provision, yeah, build and deploy. So, yeah, if we think that it can be covered in these two then it's fine.

Swarup Donepudi  18:58  
This stuff is somewhat closer to this story. Obviously, planter a is a is our sales tool. So that cannot, you cannot map any of its actually going to be applying

Speaker 1  19:11  
to all, all of these, yeah, but we are not explicitly mentioning Yeah. This

Swarup Donepudi  19:16  
only falls for,

Speaker 1  19:19  
also falls in the same category, where it has all these categories. What? Okay, what about self

Swarup Donepudi  19:27  
service DevOps, again, falls for everything. Service hub is for Bootstrap, build and deploy and maybe discover. Discovery, something that we are missing, which is on internal developer platforms. One of the core values that you get is you get your new developer. Sorry, onboarding new developers.

Suresh Attaluri  19:45  
Your discover. Discover is

Swarup Donepudi  19:49  
you want to know, what are all the list of services your organization has? Who is the owner of that? Where is it being deployed? What is the health of those? Etc, etc. Is all the information put in a central place for you to discover. So that is our service.

Okay? So that is where we are saying bootstrap will also

Suresh Attaluri  20:17  
discover the Discover point, I think a bootstrap

Swarup Donepudi  20:19  
Build and Deploy. Yeah, here we are not, do we? I don't know how many it is going to get to. Okay. So far we have provision Bootstrap, and maybe we can combine some of these also provision discover, but

Speaker 1  20:38  
we can all we can prepare the list right we can always exclude, yeah,

Swarup Donepudi  20:43  
so, so we can add discover to this train right now. And this is related to allowing either new team members or existing team members to learn about the all the services that exist within the organization, etc. That is what we call it as service hub, which is streamline your Bucha, creation, all your services, one hub. So this is the vocabulary,

Speaker 1  21:12  
but that will also discover will also our quick action functionality. Using our quick action functionality, users will be able to discover what is the ready. I mean, if there is a Redis instance, it's not only applicable for micro services.

Swarup Donepudi  21:32  
In fact, I recently thought about our own like thought of service hub, which doesn't only have to show the make these services. It can include these also, like this is also a service in a way. So, yeah, we can. So what is it that you're proposing here? No, I'm not,

Speaker 1  21:57  
I'm not proposing to modify anything about service. Okay, I'm saying when we talk about this, you can services

Speaker 2  22:06  
both Microsoft service. So I

Swarup Donepudi  22:10  
think it doesn't have to be by quick launch. It can actually be within service. Hub is, is my new proposal, which is we thought we will only do when it goes, go to service. It's all like in house built services. But you can also explore, oh, there is a Redis service running, and it is running, say, on Google memory, store, whatever. And here is who created it, but that's a service, but that actually probably, I think we have we added an annotation or an option to our API resource, kind to say, is service right? So we only did for micro service resource kind fargate, cloud run or something. Maybe we have to extend that to other places like Redis, Kafka. We need to identify them and surface them in service, right? Yeah, okay, so that's an action, but Okay, going back to these options, provision, bootstrap builder, deploy, operate, discover, monitor, monitor and

Speaker 1  23:21  
anything about I am, we can add here like,

Swarup Donepudi  23:26  
so we didn't claim that on the landing page, okay? It goes on the Features page the same so, and that also is put at the end. We can reorder this. Yeah. So, yeah, Team administration, management, blah, blah, blah, but yeah, again, I think these are, like, functional. It's not,

Speaker 3  23:48  
it's not, like, it's not the main, yeah,

Swarup Donepudi  23:55  
you don't, you don't do just a permission management for the sake of accomplishing your goal of building software.

Speaker 1  24:02  
This is like, like what we said initially, the theme is what you can do on brand and cloud

Swarup Donepudi  24:09  
and that that will have a direct value proposition, like, provision has a direct outcome. Bootstrapping has an expectation, has a result. Billion deploy has an expectation as a result, but I am there is no clear expectation, like, what is it that you want to do in this process

Suresh Attaluri  24:27  
that's not the core functionality

Swarup Donepudi  24:30  
or value that you graduate you will need it just like login or sign up. You will eventually need proper ability to control or manage access between your resources. But that's not like I need it, without which I can deploy my software or make my software available to users. So

Suresh Attaluri  24:49  
anything else, if we want

Swarup Donepudi  24:51  
to put it, maybe it can have something like manage access or something, because I'm trying to build in words here to so manage access.

Speaker 1  25:05  
Can an agent team or managing access? Okay, you can manage your team. That's always

Swarup Donepudi  25:10  
manage teams is not what is the goal we are even we are offering, right? We provided teams to enable access management, okay,

Speaker 1  25:23  
yeah, managing access. Of these, manage access, manage access. Cool. I think that's

Swarup Donepudi  25:32  
that's a good discussion. I think another 50 minute long discussion, and I'm pretty sure there is so much notes that can be covered, created, and I think we are still at the hero section, where we are we decided on one of the important aspects, what do you want to show the users on their first visit? That's a pretty big deal. And in the next discussion, we'll move forward with talking about each of these core features that we added here, yes, and talk about what, what are those? What is the media that we want to show it on, on the landing page so it doesn't have to be expensive, or maybe, maybe there will be a video where these, these don't have to be like graphical. The first parts will have to be graphical, whereas, as the user kind of starts to go down the page, we can just have a video where we are explaining our own app, what it does, right? And all of these, at least, are capable of demonstration today, except service of Bootstrap, launch. Everything else, we have a lot to show. So in the next section, we'll take one feature at a time, talk about, talk about, what should go on the landing page for that like, so what is it that you put a landing page is the overall value proposition of that particular feature, but each feature has its own dedicated this thing. Okay.

Speaker 1  27:20  
Yeah, so two things, what has to be there on the landing page under feature and

Swarup Donepudi  27:25  
yeah. And then there is a dedicated page for you. So and the hero element will may have the same video as as the media, but now going going for each of these, everything has a section title and section description, and here we want to show. We don't have to show a lot of videos. Now, if ever we feel like, okay, this this section, we can show something from the app, and that can be simple recording, but the goal of our conversation is to finalize those things, yes. So decide whether or not we want to show anything from the app at all, if it is a simple screenshot, or if it is from figma. Put it like, for example, high cost of waiting. This can also be just an animation, right? There is nothing on our app to show high cost of waiting. It's a story or value problem that we are saying. So this can be an animation like illustration outside graphic, and unleash the power of self service. So we say this thing again, do we want to put an animation here, or do we want to show something on our app? Recording stuff from our app is the easiest way to add media Everything else requires extra work. So we'll keep that in mind and we'll finalize. So we will do one feature for session going forward. Yeah. Okay. All right, I'm concluding the session here, it's another good discussion.

