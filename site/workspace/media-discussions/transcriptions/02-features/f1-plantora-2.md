Swarup Donepudi  0:10  
Okay, here we are another good, good morning. Okay, so with

Speaker 1  0:21  
our new setup, with the new setup, yeah, that's the exciting part.

Swarup Donepudi  0:25  
We have this new mic. Sennheiser, M, K, H, so we are hoping that the audio quality is going to be good, and sure she is more excited about about the new, new devices that have been added to the studio. And the purpose of this discussion this morning is so we were actually doing a video series on discussing all the aspects of with the features so with a goal to add media or create media. However, when we were talking about the feature page for plant or itself on on our design, we kind of like realized that this design does not include the new features that we have added after I handed off the requirements to or after I finalized the sections on this page, and I'm going to first open up The console app and go over what those features are here with Suresh, and we are going to essentially talk about why we added those features and what value anyone is going to get out of those features. And then use this, use the transcription or the use the this conversation as input to chat GPT to come up with or add more section to plantora feature page, and maybe we'll talk about the media for those features. Well, maybe we are going to probably cover that in the video itself.

Suresh Attaluri  2:20  
Right? I think we can talk about the journey, like, before showing it on the platform, okay, we ended up building that feature, like, Why? Why? Is a problem that we faced? Yeah. And maybe we

Swarup Donepudi  2:33  
should also talk about the feature that we already planned but haven't implemented it yet, which is the infra job summary, it's not it's right on the board. It's like, we want to leverage the same APA resource chats to also do infra job summary and have the provide the ability. So we dive into those duties. And just trying to remind that, because I just thought about it. So we started off this discussion, we only talked about sharing chats and ABA resources level chats, but infra job summaries and ability to converse on the infra job failures right within the chat is something that we plan in the pipeline. So we'll also talk about it. So yeah, where do you want to begin? Like, what? Why do you what prompted us to, okay,

Suresh Attaluri  3:23  
so those things, so we will our plant, the Chatbot, the Chatbot, yeah, and we're doing things at different levels, like we're provisioning resources, we're Testing and all we're doing all those stuff, but those charts, in a way, are isolated and only limited to the person who has requested the or created those charts. Yeah, and there's no aspect of collaborating with other team members, yeah, and we felt that collaboration is a thing that is needed, yeah, so that multiple peoples can put on their request in a single chat. Yeah, that's the whole intent. And

Swarup Donepudi  4:13  
for me, the what actually prompted or felt like this could make sense. Was every time I ran into some issue, I wanted, I was trying to create screenshots of that conversation and I was sharing with you same thing. I thought same thing would be the case even within the future customers of us, where one developer is trying to do something in their chat and and they want to share the conversation with another fellow developer. And I always felt like chatgpt itself should have added this feature, because I always wanted to collaborate with our team on chatgpt. So since we have the opportunity, and I felt like it was a low hanging fruit, so it is both like value adding and also easy to roll out for us, so that that is when I

Suresh Attaluri  5:07  
also want to highlight the we have not we have identified two levels of collaboration. One is a user intentionally sharing a chat, yeah. And there is another level of collaboration, which is at the resource level, yeah, where anything happens at the resource level? Chat,

Swarup Donepudi  5:26  
yeah. So when you say resource, you're talking about any multi cloud deployment,

Suresh Attaluri  5:30  
deployment. AP, yeah, infra component, any deployed resource, yeah, yeah, will have a chat associated with it, okay? And any, anything, any request put over there will be communicated or will be shown to sorry.

Swarup Donepudi  5:53  
Everyone can access conversation. It's not intentionally, anyone sharing anything, the very default nature of those conversations, they begin right close to where the deployment configuration can be found on the console app, and the conversation is going to live there, so anyone who has access to that configuration can also Explore historical conversation that happened on that resource. Right? So

Suresh Attaluri  6:24  
users of those charts, I mean resource levels, specific chairs should be aware that whatever is put over here will be broadcasted to all the users who have access to that that resource. So that's, yeah, this share chat feature is more individual preference.

Swarup Donepudi  6:46  
I noticed another difference between these individuals starting conversation process, the conversations that happen at the APA resource level or the deployed resource level, is in order for the resource, in order for deployment to happen, say, for example, I want to deploy a Redis Database on Google Redis, before that creation deployment happens, that deployment doesn't exist, right? So it's not possible to have a conversation on something that doesn't exist. So the creation workflows may happen or begin at individual, individual level once the resource, once something gets deployed, now the configuration exists, and during the life cycle operations, or during the life cycle of that deployment, any configuration of deaths can live close to that. So that is one of the observation that I made, and another one is, yeah, another observation was, whenever you begin a conversation about any deployed resource, say you want to update the configuration or you want to rerun an IAC stack for that particular resource, for whatever reason, etc. First of all, you need to give provide the context to the Chatbot about which deployed resource you want. You want to you're talking about, right? But when you go and start the conversation right at the deployment, deployed resource, yeah, you don't have to provide the traditional

Suresh Attaluri  8:20  
context that's already there. Yeah, that's already there. That's the difference, yeah.

Swarup Donepudi  8:24  
So you directly go and express your intention, as opposed to, if you are doing the conversation as an individual, first of all, you, let me say, hey, I want to do something x on some resource. So bot needs to First, identify, look it up, identify, confirm, so all of that can be avoided if you the depth collaborate,

Suresh Attaluri  8:44  
right? Yeah, as a platform owner, I believe we would recommend to use this resource level charts for all resource level operations, rather than individual charts. But we are not

Swarup Donepudi  8:57  
restricting them together. It's a recommendation,

Suresh Attaluri  9:01  
because you don't need to set the context. You don't, and the broadcasting part you it is automatically said to everyone. Everyone is informed looking at the chart, but when it comes to individual chart, you hide certain details, like people won't be informed, yeah, is they will have the version history and other features, but in short, they might not be, yeah, that's the switch.

Swarup Donepudi  9:29  
Let's see how the audio is going to change. Yeah, okay, so now that we went over, let's see that in action. Is that fine?

Unknown Speaker  9:40  
Yes, I think

Swarup Donepudi  9:41  
we talked about the whys. Now let's see the how. So, yeah, so individual conversations can be explored from ask by clicking on Ask plan, and this is where you see both the chats that I started, the conversations that I began. And this icon indicates the chats that I shared with others. And these are the chats that I am a guest, meaning somebody invited me to collaborate, right? So if I scroll back up, the first message is always a good indication of who started it, because that's the person moving inside. So in this case, is most likely you, I guess, or Satish. This is Satish, yes, and this is you, obviously so. And as you can see, because this is a collaborative, shared chat, I was able to also, it's important to understand that any conversation that happens in this collaboration, or even at the resource level chats, every single message is, is going to be processed or responded by what, no matter what. So the these are, these chats are not meant for, like, regular conversations that are that happen typically in a Slack channel, right? Yeah. So,

Suresh Attaluri  11:04  
whether anything, I mean, even though, if a single chart is shared across multiple people, if they are adding some message to the chat, yeah, they're actually talking to the bot, not to not with the person who is with whom the chart is shared, yeah? Like, that's a whole. I mean, we don't see that as a feature, right? We don't see that is required, like, if, if we both are in a single chat, and I want to address you and say something to the person. This is not the place for you to put that. This is purely to talk to

Swarup Donepudi  11:51  
the board. Yeah. And sharing the chat is pretty straightforward. Wherever you're talking, you can look up the member that you want to share the chat with and there is the sharing is not limited to organizations, or is it? It is all it is. Share chats with other members in your own organization, but not others. Yeah, so depending on which organization you it's in the context. But it doesn't matter if the environment, is the context, right?

Unknown Speaker  12:24  
Yeah, I think I should check that. But

Swarup Donepudi  12:25  
ideally, you can be able to do it with anyone

Suresh Attaluri  12:30  
in the organization only, yeah? Sharing thing, okay, yeah, okay, the chat can only be shared to. Another member who's in the same organization.

Swarup Donepudi  12:46  
Yeah, you can develop. I can type in the email address of anyone. Yeah, for example, if I want to share this chat with Satish,

Suresh Attaluri  12:57  
yeah, you're getting Satish email id recommendation is because they belong to this. Yeah.

Swarup Donepudi  13:03  
Also, we forgot about the teams aspect, yes. So you can share the chat with the entire team without having to invite each so this is very powerful feature to me, like if you want to collaborate with and bring everyone on your team. So that's, again, very powerful addition to the mix. So yeah, that's, that's, that's, that's good. So let's now go to the resource level chat. So I am on an AWS DynamoDB table that I deployed yesterday. And yeah, I can do ask plantora, and I can go back to the history of this conversation. And as you can see, I already made a few changes. Not even made a few changes. I started by asking, Why is the last time job failing? I saw that, and it says, I need the infra job ID. I said, Can you look it up from the details? And it says, I currently do not have the capability to look up in saddle. I said, the infra job ID is in straighter section of the API resource. And it did look it up and and then I said, Okay, now that you know, I asked, Did you find it out? By the way, it says, Please hold on for a moment. And I was, I was surprised that it would come back with another message. But that doesn't happen,

Suresh Attaluri  14:31  
right? Oh, that doesn't happen. I don't know why it is asking, yeah, did you find out?

Swarup Donepudi  14:39  
And yeah, I just waited for a couple seconds, and I asked you to find out, and then it actually came back by looking up the details of that infra job ID and figured out from the error messages that, oh, this is what is causing the failure, and here is how you can fix it. So overall, as like, what we are trying to discuss here is the advantage of resource level chance, and anyone who has access, which can be discovered by going to the permissions management and we say, Show inherited permissions, and yeah, like All these different people have access to this resource, yeah, so the

Suresh Attaluri  15:23  
organization level? Yeah, that is where we they can, since they have a organization, but they can see that, yeah, this

Swarup Donepudi  15:32  
is probably a discussion for a different Yeah. But is this intuitive? Is my question. We'll talk about it when we talk about IAM, but going back, yeah, overall, anyone who has access to just look up the detail you permission to say resource can access this entire chat, and they can start the conversation, and what they can perform on this resource totally depends on what kind of permissions that particular identity has on that resource. So just because you have access to this chat doesn't mean you can deploy. You can undeploy this resource. You need to have appropriate permissions to on deploy. So if you if you come here and say and deploy, and if you don't have permissions to do that, the system will deny that request automatically.

Suresh Attaluri  16:22  
The beauty of this individual charts is now I can continue the chart where you have left, like, figure out the root cause. Now I it's not required for me to start the whole chart from the beginning, yeah. I can start where it was left, and I can try to fix this issue, right? Yeah,

Swarup Donepudi  16:41  
yeah. This is definitely another powerful feature. So now we move on to the third aspect of this discussion that we wanted to talk, which was, Okay, interesting I'm unable to navigate to the that's a Bucha. Okay, so another area where we have already planned and we haven't actually rolled out this additional collaboration aspect of these chats is as soon as whether a pulumi or a telephone stack execution is completed, our system will automatically do a summary of what happened as part of that Stackdriver, and make that as a summary for here and the developers, or anyone who is who is a stakeholder of that particular deployment, they can start conversations on the summary itself, and we, we are, those are that is also essentially an API resource level chat, the APA resource here being the infra job itself,

Suresh Attaluri  17:53  
yeah. But this has to be handled somewhat different, yeah.

Swarup Donepudi  17:58  
Experience is going to be different, as opposed to having a conversational deployed resource versus this, these are not at resource level. These are at the infra job level. We are the scope. The chart is only scoped for that particular stand job. Yes.

Suresh Attaluri  18:15  
So initially, the only difference that I see is initially, when we create a chart for AWS DynamoDB resource as as a platform, we are doing nothing, nothing, yeah, but when we talk about infra job, yeah, so we are triggering some analysis

Swarup Donepudi  18:32  
to follow up on your thought of like, who, when is that conversation on an API resource or a deployed component begins is when the first message is like, the developer will start. The conversation will always begin when a developer wants to have a conversation, whereas with infra job, it's a asynchronous,

Suresh Attaluri  18:57  
the initial conversation will be started by the platform

Swarup Donepudi  19:01  
plan control the back end, and then the developers can continue that conversation by asking more questions around the summary itself. Yes, and we want to make these tag job summaries and the follow up conversations to be far more effective by making or by enabling the chat bot, or by providing chat bot with all the necessary context to provide as much relevant response as possible. And when I say more context, I'm talking about all the input that was used to run the infra job and the pulumi or TerraForm code that was used to execute that infra job as input, the system has all of that information, but we need to enable the bot. We need to build that capability into the bot to bring all of this information so that the bot can provide will have more context. For example, if it is an error, then bot will look at the input. Bot will look at the COVID or TerraForm code, and it has a better chance of identifying what could have caused that failure. Yes, so I believe that is another huge value that we can provide to the developers, because deployment values are one of the biggest frustration causing areas, and we have an opportunity to solve it very effectively. So I think that definitely concludes

Suresh Attaluri  20:35  
our I think this feature also comes under that operate part, because I'm trying to deploy something.

Swarup Donepudi  20:43  
Operators, again, more more or less is post deployment. Okay, so

Unknown Speaker  20:47  
this is

Swarup Donepudi  20:53  
more of a deployment support, yeah. So, I mean, typically, like walk, who would have solved this is developer would reach out to a DevOps engineer or somebody who has created the module, and then they'll say, Hey, I ran the job and it is failing. I don't understand. In most cases, developers don't or they are not approved with the right knowledge to be safer. Those other messages and get to the root cause of it. So it's usually the engagement with the DevOps or support engineer. So that's that's also a lot of bottleneck. They wait for somebody to look at her. I've seen this first time at CDP, and then something fails. They ask me, I walk through my queue. So they wait until I get to that. So all of this is avoided with real time. Do you have any more concluding comments on on these particular aspects?

Suresh Attaluri  21:50  
No, I think okay, so I think we need to discuss how this whole feature is going to be shipped out on

Swarup Donepudi  21:58  
the landing page. Yeah, I think all of these conversations were so, yeah, this concluding notes is for chatgpt, and essentially, I'm going to share the existing sections of this finalized plantora Features page, and I'm going to share the transcript of our previous conversation, where we talked about the media for each of the sections, and then I'm going to share the transcription for this conversation where we talked about the capabilities that we have added after the previous page was finalized, And the expectation from chatgpt is to add more sections to the page talking about these three different aspects and and also, I think this page is missing the important of the important aspects. I think it does make simplified permission to access management is probably, I think I went to discuss, or maybe I don't know. So overall, one of the aspects of this chat conversation is that it's the bot is built in such a way that,

Suresh Attaluri  23:20  
until now, when we talked about in this space, yeah, we talked about the intents that plant dollar chart can solve, yeah, like when we talked about the guided generation and access management is, it's more about what you can do with this thing. Yeah, this feature sharing charts is, I don't know. It's not about intent anyway. This is, it's more

Swarup Donepudi  23:45  
like providing more value

Suresh Attaluri  23:50  
to the how you can collaborate with

Swarup Donepudi  23:52  
collaboration is, yes, and one of the aspects that I find missing on this page is that all the chat conversations are I am protected, meaning you can ask whatever you feel like asking, but bot will ensure that it will only execute the steps that you are authorized to execute. Yes, if you have get access on any resource, and if you would like to look up the details, bot would show that information. But if you don't have update permissions on any resource, even the guided like the update workflow will not work, because, again, you will not have

Suresh Attaluri  24:37  
access to that. So I believe chat GPT is going to give the section, section title some subtitles for this new feature. But I think we still need to figure out the media, media, right?

Swarup Donepudi  24:53  
So media is again going to be, oh yeah, I think the first section we talked about two videos that we want to say. One is access provision. The first part is provision. Second one is access management. And I think I can cover the this. I am permissions aspect in access management itself, like there are various parts of it, which is platform engineers sharing connections with environments. They can do it from here, okay, and members can add other members to their team from the chat, unrelated to provisioning, but still related to access management, and then also highlighting the fact that the board is always aware of the permissions that the user has to perform any operation that's and we are going to add third sections in collaboration.

Unknown Speaker  25:50  
Yeah, third video for collaboration, collaboration. And

Swarup Donepudi  25:53  
do you think we should add a fourth video talking about the value for deployments assistance, like infra job insights or that's also collaboration in a way, but it also has the added benefit, which is not

Suresh Attaluri  26:09  
applied To API resources like we discussed earlier, also true. It falls under the category, yeah,

Swarup Donepudi  26:24  
maybe we can record that video and put it on both pages. That's also one option.

Unknown Speaker  26:32  
We'll see, we'll see, yeah.

Swarup Donepudi  26:33  
But overall, I think it falls in both Yeah, but

Suresh Attaluri  26:38  
as media fit, we will have a video,

Swarup Donepudi  26:40  
yeah, I think you're right that should actually go on that page. They integrated IAC workflows or whatever, because that's more like we are giving you a workflow, and it runs instantly. You get to see the days all that. Yeah,

Suresh Attaluri  26:57  
when we talk about chat, the power of chat is to understand your reading letters and deliver, deliver, yeah, and making you have that less learning curve or less Yeah,

Swarup Donepudi  27:09  
yeah. Maybe we talked about the provisioning part in the beginning, right? Maybe we can add a line or two saying, yeah, if you the summary, start your phase, you you get the summary. Will just mention it, but will not go dive deeper into it, meaning we don't have to make a separate video for it. You can include of doing the provisioning partners, yeah.

Suresh Attaluri  27:32  
The theme of this feature should be more of chat, chat, where user is expressing even their intent or request,

Swarup Donepudi  27:43  
what to do, collaborating with the how is

Suresh Attaluri  27:45  
the chart? Responding? Yeah,

Swarup Donepudi  27:48  
I would, I would start to call plant or collaborating with or engaging with the developer, because we are saying that right, like either platform engineer ready to help 24/7, so I'm thinking plant a as a persona, as a chat bot. So we can say developer expresses his intent to plant Ara, and plant Ara will work with the developer to get to make sure he gets what he needs. Yeah, I think cool. So

Suresh Attaluri  28:24  
we decided that we'll going to add a third collaboration part of chat. And, okay, yeah, let's discuss the script, also one script, then we can discuss the the media is at the bottom before

Swarup Donepudi  28:42  
we cannot talk about the media without the sections finalized. Okay, those sections and the subtitles are required, so we'll only talk about the video and we'll talk about the sections also finalized. So yeah, what do you think might make a good two or three minute video showing off the collaboration aspects of chat feature.

Suresh Attaluri  29:07  
We'll start from that, that part where we have in the provision video we might have created, we had a script of creating some resource, right? We'll continue. We'll continue from there. Hey, now I want to share this. I want, I want a fellow developers opinion on whether what, what should be the right value here, yeah, so I'll invite him to this chat, and we both are an individual chat, yeah,

Swarup Donepudi  29:32  
that would be an individual. Yeah, you're still creating it.

Suresh Attaluri  29:39  
We can do it for a individual or a team? Yeah,

Swarup Donepudi  29:44  
I think we need to explicitly call out that aspect so that. So I want chatgpt to know that we want to include that highlighting the aspect that you can I can invite either my fellow developer, or I can invite my entire team using the notion of teams on random cloud, and they can come and collaborate. Now we'll show, we'll have

Suresh Attaluri  30:07  
two sections of users. Yeah, we'll show the other guys. Now I want to see, I

Swarup Donepudi  30:12  
think that's probably going to be tough

Suresh Attaluri  30:15  
to record. Yeah, Okay, how about we have two screens side by side, half of the screen being user one, half of the screen being user

Swarup Donepudi  30:25  
two, yeah, I think the icon login has two different users. Yeah, I can do that, and

Suresh Attaluri  30:30  
we'll show that this chat is not yet shared. Now it is shared. Yeah,

Swarup Donepudi  30:35  
for that, I need to fix a bug, like when two sessions are open, there is a bug that is preventing messages. I need to write it out, but that will force me to fix it as, yeah, that's a good demo. I

Suresh Attaluri  30:53  
want to invite that. This will be the individual part of the sharing, yeah,

Swarup Donepudi  30:58  
and then the same

Suresh Attaluri  31:00  
resource. We'll jump on to individual API, resource level chat. We'll try to modify something and see if the other users, yeah, they refer to the

Swarup Donepudi  31:14  
page, or they go to the same chat, and they see that I have just made some modification, and he gets to see that we

Suresh Attaluri  31:23  
can, we can do this thing like user one is requesting for a modification. Yeah, plan tour will ask for final confirmation that this is the change. That's how the guided experience is. And then the confirmation should be given by the second user to do the journey. So then

Swarup Donepudi  31:41  
I mentioned, while recording the video, that I pinned my developer on Slack, asking for his contribution. So he jumped on to the API resource and looked at the chat. He reviewed it. Now he says, proceed, and the body is going to proceed with the next steps. I think that's a good small script that can explain the overall capability there. Yes, and then

Suresh Attaluri  32:08  
nothing. We should also, okay, this collaboration will put it in the individual level chat only. Sorry, API resource level chats. We won't do that multiple thing in the individual, yeah. I think individual chat is good enough. When someone shares, the other person is able to see the chat, see the chat there, ends the individual

Swarup Donepudi  32:29  
there. Maybe we can simply say, Maybe we should know, change, like, set the CPU limit to, like, 400 rather 300 just one message, just to say they can collaborate on

Suresh Attaluri  32:47  
the user too. Yeah, with whom the chat was shared, they will. Yeah, I think

Swarup Donepudi  32:54  
we'll see that doesn't matter. I guess whatever we are already showing is definitely a step improvement. So people, as long as they understand the overall collaboration, they it makes sense that they can, we can only say that the now they can collaborate too. So it's okay to, like, not show everything you skip the infra jobs, collaboration here. What do you think? Yes, yeah, it's not. Collaboration is not the key, right there is it's going to be around the developer doing something to resolve his own deployment. But as per, the team has access to the chat, but it doesn't. I think we can sell a stack

Suresh Attaluri  33:36  
job level. It's less of a collaboration resource is infra job is life of charge less, right? Because a new infra job, we can be triggered and the a PhD resource state gets changed. So I believe infra job charts will have less conversations over there. Yeah, but more of analysis, friend,

Swarup Donepudi  34:02  
explain. Successful. Nobody even looks at what happened,

Suresh Attaluri  34:08  
right? Yeah, they just expect to consume some information or what is the issue and how can they fix it? That's what. And rather than putting more collaborating, adding more messages or having any more conversations,

Swarup Donepudi  34:23  
I think that's a good 35 minute discussion. We think this is all sufficient for me to get to finalize the additional sections with their titles and subsections, and also the script with all the steps. And once the script is ready and once we get the subsections, we'll probably take just screenshots or gifs of the same showing the same stuff, and put it next to these subsections. Awesome. That's a great discussion. Thanks for collaborating once again. Suresh, you have always been a great

Suresh Attaluri  34:56  
teammate. Likewise. Thank you likewise, much. Do

