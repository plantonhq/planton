/**
 * What customers said, word for word, attributed to the person who said it.
 *
 * Each quote is copied verbatim from its source of record in the company
 * repository (company/sales/customers/<customer>/README.md), where the
 * person's approval for public use is recorded. Nothing here is paraphrased,
 * tightened, or attributed to a company instead of a person, and no quote
 * carries a dollar figure. A quote whose approval is not on record does not
 * appear here.
 */
export interface Testimonial {
  name: string;
  role: string;
  company: string;
  location?: string;
  quote: string;
  /** Original customer portrait, retained from the earlier public site. */
  avatar?: string;
}

export const TESTIMONIALS: readonly Testimonial[] = [
  {
    name: 'Rohit Reddy Gopu',
    role: 'CEO',
    company: 'TynyBay',
    location: 'India',
    quote:
      'For one client project where the client mandated GCP but our DevOps engineer had no GCP experience, Planton allowed us to successfully deliver the entire infrastructure. We essentially got full DevOps capabilities for GCP without needing GCP expertise on our team.',
  },
  {
    name: 'Sai Saketh',
    role: 'Junior DevOps Engineer',
    company: 'iorta TechNext',
    location: 'India',
    quote:
      'As a junior DevOps engineer with almost no AWS experience, Planton enabled me to provide a very mature developer experience to our entire 7-person dev team. They can quickly deploy services to multiple environments without me having to deal with learning AWS from scratch or rewriting complex infrastructure code.',
  },
  {
    name: 'Balaji Borra',
    avatar: '/_site/images/customers/people/balaji-borra.png',
    role: 'DevOps Engineer',
    company: 'TynyBay',
    location: 'India',
    quote:
      'Planton has dramatically improved my efficiency. I no longer have to deal with the grunt work of rewriting Terraform configurations between client projects. I can now manage multiple client environments simultaneously and provide a much better experience for all the developers I support.',
  },
  {
    name: 'Rakesh Kandhi',
    avatar: '/_site/images/customers/people/rakesh-kandhi.jpeg',
    role: 'Senior Developer',
    company: 'TynyBay',
    location: 'India',
    quote:
      "The dot-env file generation for services feature in Planton's ServiceHub been super helpful for me. I can now update service configurations without having to ping Balaji every time. Even better, creating new services and deploying them to dev, staging, or prod is completely self-service. I don't need to wait for DevOps anymore.",
  },
] as const;

/** Look up a quote by the person's name; throws at build time if a page names someone who is not on record. */
export function testimonial(name: string): Testimonial {
  const found = TESTIMONIALS.find((t) => t.name === name);
  if (!found) throw new Error(`no testimonial on record from "${name}"`);
  return found;
}
